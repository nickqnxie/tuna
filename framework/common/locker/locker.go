/*
@Time : 2019/12/14 13:11
@Author : nickqnxie
@File : locker.go
*/

package locker

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"tencent.com/lock-cron/handlers/monitoring"

	api "github.com/hashicorp/consul/api"
)

const (
	// defaultLockRetryInterval how long we wait after a failed lock acquisition
	defaultLockRetryInterval = 10 * time.Second
	// session ttl
	defautSessionTTL = 30 * time.Second
	// if can't acquire, block wait to acquire
	defaultLockWaitTime = 5 * time.Second
	// block min wait time
	minLockWaitTime = 10 * time.Millisecond

	TryAcquireMode = iota
	CallEventModel
)

var (
	ErrKeyNameNull = errors.New("key is null")
	ErrKeyInvalid  = errors.New("Key must not begin with a '/'")
)

// DisLocker configured for lock acquisition
type DisLocker struct {
	isClosedDoneChan int32
	doneChan         chan struct{}

	consulLock   *api.Lock
	IsLocked     bool
	ConsulClient *api.Client
	Key          string
	SessionID    string

	LockWaitTime      time.Duration
	LockRetryInterval time.Duration
	SessionTTL        time.Duration
}

// Config is used to configure creation of client
type Config struct {
	Address           string // consul addr
	KeyName           string // key on which lock to acquire
	Token             string
	LockWaitTime      time.Duration
	LockRetryInterval time.Duration // interval at which attempt is done to acquire lock
	SessionTTL        time.Duration // time after which consul session will expire and release the lock
}

func (c *Config) check() error {
	if c.KeyName == "" {
		return ErrKeyNameNull
	}

	if strings.HasPrefix(c.KeyName, "/") {
		return ErrKeyInvalid
	}

	return nil
}

func (c *Config) init() {
	if c.LockRetryInterval == 0 {
		c.LockRetryInterval = defaultLockRetryInterval
	}
	if c.SessionTTL == 0 {
		c.SessionTTL = defautSessionTTL
	}
	if c.LockWaitTime == 0 {
		c.LockWaitTime = defaultLockWaitTime
	}
	if c.Address == "" {
		c.Address = "127.0.0.1:8500"
	}
}

func NewConfig() *Config {
	c := &Config{
		LockRetryInterval: defaultLockRetryInterval,
		SessionTTL:        defautSessionTTL,
	}
	return c
}

func ConsulPut(client *api.Client, key string, value []byte) (err error) {

	p := &api.KVPair{Key: key, Value: []byte(value)}

	kv := client.KV()
	if _, err = kv.Put(p, nil); err != nil {
		logrus.Error(err)
		alarm := new(monitoring.Alarm)
		ip, _ := alarm.LocalIPv4s()
		alarm.Content = fmt.Sprintf("consul 集群连接失败，服务无法正常启动. err: %s", err)
		alarm.ObjName = ip
		alarm.PushAlarm()
		return
	}

	return nil
}

//加载备份consul集群配置
func backupConsul() (client *api.Client, err error) {

	return
}

// New returns a new dislocker object
func NewLocker(o *Config) (*DisLocker, error) {
	var (
		locker DisLocker
	)

	// init config
	o.init()

	// set consul server address
	cfg := api.DefaultConfig()
	cfg.Address = o.Address

	if o.Token != "" {
		cfg.Token = o.Token
	}

	logrus.Debug(cfg)
	// instance consul client, new client share http.conn in DefaultPooledTransport
	consulClient, err := api.NewClient(cfg)

	//测试consul是否能正常工作
	if err := ConsulPut(consulClient, "checkkey", []byte(`ok`)); err != nil {
		return &locker, err
	}

	if err != nil {
		logrus.Errorf("new consul clinet failed, err: %v", err)
		return &locker, err
	}

	// set
	locker.doneChan = make(chan struct{})
	locker.ConsulClient = consulClient
	locker.Key = o.KeyName
	locker.LockWaitTime = o.LockWaitTime
	locker.LockRetryInterval = o.LockRetryInterval
	locker.SessionTTL = o.SessionTTL

	if err = o.check(); err != nil {
		return &locker, err
	}

	return &locker, nil
}

// RetryLockAcquire attempts to acquire the lock at `LockRetryInterval`
func (d *DisLocker) RetryLockAcquire(value map[string]string, acquired chan<- bool,
	released chan<- bool, errorChan chan<- error) {
	ticker := time.NewTicker(d.LockRetryInterval)

	for ; true; <-ticker.C {
		value["lock_created_time"] = time.Now().Format(time.RFC3339)
		lock, err := d.acquireLock(d.LockWaitTime, value, CallEventModel, released)
		if err != nil {
			logrus.Error("error on acquireLock :", err, "retry in -", d.LockRetryInterval)
			//consul集群不可用发告警

			alarm := new(monitoring.Alarm)
			ip, _ := alarm.LocalIPv4s()
			alarm.Content = fmt.Sprintf("error on acquireLock : %s retry in -%v", err, d.LockRetryInterval)
			alarm.ObjName = ip
			alarm.PushAlarm()
			errorChan <- err
			continue
		}

		if lock {
			logrus.Debugf("lock acquired with consul session - %s", d.SessionID)
			ticker.Stop()
			acquired <- true
			break
		}
	}
}

func (d *DisLocker) tryLockAcquire(wait time.Duration, value map[string]string) (bool, error) {
	locked, err := d.acquireLock(wait, value, TryAcquireMode, nil)
	if err != nil {
		logrus.Error("acquireLock failed, err: %v", err)
		return locked, err
	}

	if !locked {
		logrus.Infof("can't acquire lock, session: %s", d.SessionID)
		return locked, nil
	}

	d.IsLocked = locked
	return locked, nil
}

// TryLockAcquire
func (d *DisLocker) TryLockAcquire(value map[string]string) (bool, error) {
	return d.tryLockAcquire(d.LockWaitTime, value)
}

// TryLockAcquire
func (d *DisLocker) TryLockAcquireNonBlock(value map[string]string) (bool, error) {
	return d.tryLockAcquire(minLockWaitTime, value)
}

// TryLockAcquireBlock
func (d *DisLocker) TryLockAcquireBlock(waitTime time.Duration, value map[string]string) (bool, error) {
	return d.tryLockAcquire(waitTime, value)
}

func (d *DisLocker) ReleaseLock() error {
	if d.SessionID == "" {
		logrus.Debug("cannot destroy empty session")
		return nil
	}

	defer func() {
		logrus.Debugf("destroyed consul session: %s", d.SessionID)
		d.IsLocked = false
		if !d.isDoneChanCloed() {
			// only call once
			close(d.doneChan)
		}

		if d.consulLock != nil {
			// DELETE /v1/kv/
			d.consulLock.Destroy()
		}

		d.SessionID = ""
	}()

	// PUT /v1/session/destroy/
	_, err := d.ConsulClient.Session().Destroy(d.SessionID, nil)
	if err != nil {
		return err
	}

	return nil
}

// Renew incr key ttl
func (d *DisLocker) Renew() {
	d.ConsulClient.Session().Renew(d.SessionID, nil)
}

func (d *DisLocker) StartRenewProcess() {
	d.ConsulClient.Session().RenewPeriodic(d.SessionTTL.String(), d.SessionID, nil, d.doneChan)
}

func (d *DisLocker) AsyncStartRenewProcess() {
	go func() {
		d.StartRenewProcess()
	}()
}

func (d *DisLocker) StopRenewProcess() {
	if !d.isDoneChanCloed() {
		close(d.doneChan)
	}
}

func (d *DisLocker) createSession() (string, error) {
	return createSession(d.ConsulClient, d.Key, d.SessionTTL)
}

func (d *DisLocker) recreateSession() error {
	sessionID, err := d.createSession()
	if err != nil {
		return err
	}

	d.SessionID = sessionID
	return nil
}

func (d *DisLocker) isDoneChanCloed() bool {
	select {
	case _, ok := <-d.doneChan:
		if !ok {
			return true
		}
		return false

	default:
		return false
	}
}

func (d *DisLocker) acquireLock(waitTime time.Duration, value map[string]string,
	mode int, released chan<- bool) (bool, error) {
	if d.SessionID == "" {
		err := d.recreateSession()
		if err != nil {
			return false, err
		}
	}

	if d.isDoneChanCloed() {
		d.doneChan = make(chan struct{})
	}

	b, err := json.Marshal(value)
	if err != nil {
		logrus.Error("error on value marshal", err)
	}

	lock, err := d.ConsulClient.LockOpts(
		&api.LockOptions{
			Key:     d.Key,
			Value:   b,
			Session: d.SessionID,
			// block wait to acquire, consul defualt 15s
			LockWaitTime: waitTime,
			// if true, only acquire lock once, return.
			// if false, while acquire lock with WaitTime
			LockTryOnce: true,
		},
	)
	if err != nil {
		return false, err
	}

	// the sessionID maybe is expired or invalided ID
	session, _, err := d.ConsulClient.Session().Info(d.SessionID, nil)
	if err == nil && session == nil {
		logrus.Debugf("consul session: %s is invalid now", d.SessionID)
		d.SessionID = ""
		return false, nil
	}

	if err != nil {
		return false, err
	}

	resp, err := lock.Lock(d.doneChan)
	if err != nil {
		return false, err
	}

	if resp == nil {
		return false, nil
	}

	if mode == TryAcquireMode {
		go func() {
			<-resp
			d.IsLocked = false
		}()
	}

	if mode == CallEventModel {
		go func() {
			// wait event
			<-resp
			// close renew process
			if !d.isDoneChanCloed() {
				close(d.doneChan)
			}
			logrus.Debugf("lock released with session: %s", d.SessionID)
			d.IsLocked = false
			released <- true
		}()

		go d.StartRenewProcess()
	}

	d.consulLock = lock
	return true, nil
}

func createSession(client *api.Client, consulKey string, ttl time.Duration) (string, error) {

	sessionID, _, err := client.Session().Create(
		&api.SessionEntry{
			Name: consulKey,
			// Checks:   checks,
			Behavior: api.SessionBehaviorDelete,
			// after release lock, other get lock wating lockDelay time.
			LockDelay: 1 * time.Microsecond,
			TTL:       ttl.String(),
		},
		nil,
	)
	if err != nil {
		return "", err
	}

	logrus.Debugf("created consul session: %s", sessionID)
	return sessionID, nil
}
