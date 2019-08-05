// +build integration

package wallets_task

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/NickRI/wallets-task/db"
	"github.com/NickRI/wallets-task/domain/entities"
	"github.com/NickRI/wallets-task/infrastructure/services"
	"github.com/NickRI/wallets-task/transport/restapi"
	"github.com/go-kit/kit/log"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

var hosts []string

func getFreePort() (int, error) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	err = ln.Close()
	if err != nil {
		return 0, err
	}
	return ln.Addr().(*net.TCPAddr).Port, nil
}

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	viper.Set("driver", "postgres")
	viper.Set("dsn", "user=user password=6vdDtPeA51SGJvb host=127.0.0.1 port=5432 dbname=wallet sslmode=disable")
	viper.Set("max-open-connections", 1000)
	viper.Set("max-idle-connections", 300)
	viper.Set("conn-max-lifetime", "100ms")

	dbConn, err := db.Init()
	if err != nil {
		panic(err)
	}

	wSvc, err := services.NewWalletService(dbConn)
	if err != nil {
		panic(err)
	}

	routes := restapi.MakeRoutes(wSvc, logger)

	portA, err := getFreePort()
	if err != nil {
		panic(err)
	}
	portB, err := getFreePort()
	if err != nil {
		panic(err)
	}

	hosts = append(hosts, fmt.Sprintf("127.0.0.1:%d", portA))
	hosts = append(hosts, fmt.Sprintf("127.0.0.1:%d", portB))

	serverA := restapi.NewServer(hosts[0], routes)
	serverB := restapi.NewServer(hosts[1], routes)
	go serverA.Run()
	go serverB.Run()
	m.Run()
	serverA.Shutdown()
	serverB.Shutdown()
}

func tearUpDB(d *sql.DB) (err error) {
	_, err = d.Exec(`INSERT INTO accounts (id, user_name, balance, currency, created_at, updated_at) VALUES
		(DEFAULT, 'test1', 100, 'USD', DEFAULT, DEFAULT),
		(DEFAULT, 'test2', 100, 'USD', DEFAULT, DEFAULT)`)
	return
}

func tearDownDB(d *sql.DB) (err error) {
	_, err = d.Exec(`DELETE FROM payments WHERE account_id IN (SELECT id FROM accounts WHERE user_name IN ('test1', 'test2'))`)
	if err != nil {
		return err
	}

	_, err = d.Exec(`DELETE FROM accounts WHERE user_name IN ('test1', 'test2')`)
	if err != nil {
		return err
	}

	return
}

type response struct {
	Err  string          `json:"error"`
	Data entities.Ledger `json:"data"`
}

type results struct {
	sync.Mutex
	m map[int][]*response
}

func newResult() *results {
	return &results{
		m: make(map[int][]*response),
	}
}

func (r *results) Add(resp *http.Response) error {
	defer resp.Body.Close()
	r.Lock()
	if _, ok := r.m[resp.StatusCode]; !ok {
		r.m[resp.StatusCode] = make([]*response, 0, 100)
	}
	r.Unlock()

	res := response{}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}

	r.Lock()
	r.m[resp.StatusCode] = append(r.m[resp.StatusCode], &res)
	r.Unlock()

	return nil
}

func (r *results) ExpectCodesCount(code, count int) bool {
	return len(r.m[code]) == count
}

func (r *results) ExpectCodesErrorMessages(code, count int, errMsg string) bool {
	if len(r.m[code]) < count {
		return false
	}

	var c int
	for _, resp := range r.m[code] {
		if resp.Err == errMsg {
			c++
		}
	}

	return c == count
}

func testSendPayment(t *testing.T, w *sync.WaitGroup, users, host string, count int, res *results) {
	for i := 0; i < count; i++ {
		body := bytes.NewBufferString(`{"amount": 1}`)

		resp, err := http.Post("http://"+host+"/wallet/pay/"+users, "application/json", body)
		if err != nil {
			t.Fatal(err)
		}

		if err := res.Add(resp); err != nil {
			t.Fatal(err)
		}
	}

	w.Done()
}

func TestIntegrationDiffDirection(t *testing.T) {
	var parallel = 2
	db, err := db.Init()
	if err != nil {
		t.Fatal(err)
	}

	if err := tearUpDB(db); err != nil {
		t.Fatal(err)
	}

	defer tearDownDB(db)

	wg := sync.WaitGroup{}
	users := []string{"test1/test2", "test2/test1"}
	result := newResult()
	for i := 0; i < parallel; i++ {
		wg.Add(1)
		go testSendPayment(t, &wg, users[i], hosts[i], 100, result)
	}

	wg.Wait()

	if !result.ExpectCodesCount(http.StatusOK, 200) {
		t.Fatalf("Expect status StatusOK 200 times")
	}
}

func TestIntegrationSameDirection(t *testing.T) {
	var (
		nobalance  = (rand.Intn(25-1) + 1) * 2 // randomize results
		iterations = (100 + nobalance) / 2
		parallel   = 2
	)

	t.Logf("nobalance = %d, iterations(per goroutine) = %d", nobalance, iterations)

	db, err := db.Init()
	if err != nil {
		t.Fatal(err)
	}

	if err := tearUpDB(db); err != nil {
		t.Fatal(err)
	}

	defer tearDownDB(db)

	wg := sync.WaitGroup{}
	result := newResult()
	for i := 0; i < parallel; i++ {
		wg.Add(1)
		go testSendPayment(t, &wg, "test1/test2", hosts[i], iterations, result)
	}

	wg.Wait()

	if !result.ExpectCodesCount(http.StatusOK, 100) {
		t.Error("expect status payment sent with StatusOK, 100 times")
	}

	if !result.ExpectCodesErrorMessages(http.StatusPaymentRequired, nobalance, "test1: don't have enough balance") {
		t.Error("expect status StatusPaymentRequired " + strconv.Itoa(nobalance) + " times with message 'test1: don't have enough balance'")
	}
}
