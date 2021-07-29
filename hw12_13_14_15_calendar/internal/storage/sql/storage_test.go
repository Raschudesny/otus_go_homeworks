package sqlstorage

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/storage"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

const DSN = "host=localhost port=5432 user=danny password=danny dbname=test connect_timeout=10"

func TestSomeFunc(t *testing.T) {
	t.Skip()
	timeout, cancelFunc := context.WithTimeout(context.Background(), time.Second*20)
	defer cancelFunc()

	db, err := sqlx.ConnectContext(timeout, "pgx", DSN)
	require.NoError(t, err)
	err = db.PingContext(timeout)
	require.NoError(t, err)
	fmt.Println("Successful connect to db")

	insertSQL := "INSERT INTO events (id, title, start_time, end_time, description, owner_id) VALUES ('92061df3-0f38-4c5b-b7a2-40515ca5a514','Birthday', '2021-04-03 00:00:00 +0300', '2021-04-03 23:59:59 +0300', 'it is my birthday party time', 'f1a200f5-3f8e-4c28-b287-82376033eaae');"
	_, err = db.ExecContext(timeout, insertSQL)
	require.NoError(t, err)

	sql := "select * from events where owner_id = :owner_id"
	rows, err := db.NamedQueryContext(timeout, sql, map[string]interface{}{
		"owner_id": "f1a200f5-3f8e-4c28-b287-82376033eaae",
	})
	require.NoError(t, err)
	defer func() {
		err := rows.Close()
		log.Println("error closing sql rows: ", err)
	}()

	foundEvents := make([]storage.Event, 0)
	var event storage.Event
	for rows.Next() {
		err := rows.StructScan(&event)
		require.NoError(t, err)
		foundEvents = append(foundEvents, event)
	}

	err = rows.Err()
	require.NoError(t, err)

	require.Equal(t, 1, len(foundEvents))
	require.Equal(t, "f1a200f5-3f8e-4c28-b287-82376033eaae", foundEvents[0].OwnerID)
}

func TestSql(t *testing.T) {
	t.Skip()
	DBStorage := NewDBStorage()
	timeout, cancelFunc := context.WithTimeout(context.Background(), time.Second*20)
	defer cancelFunc()
	err := DBStorage.Connect(timeout, DSN)
	require.NoError(t, err)

	err = DBStorage.AddEvent(timeout, storage.Event{
		ID:          "03cd6323-3590-45ec-a462-4e41dcffd8aa",
		Title:       "first event title",
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(time.Second),
		Description: "first event description",
		OwnerID:     "98831c0e-c00b-43e5-840e-2f7a327ff14a",
	})
	require.NoError(t, err)

	err = DBStorage.AddEvent(timeout, storage.Event{
		ID:          "554d45c9-f8be-4de1-8152-3d4b88387055",
		Title:       "second event title",
		StartTime:   time.Now().AddDate(0, 0, 1),
		EndTime:     time.Now().AddDate(0, 0, 1).Add(time.Second),
		Description: "second event description",
		OwnerID:     "98831c0e-c00b-43e5-840e-2f7a327ff14a",
	})
	require.NoError(t, err)
	err = DBStorage.Close()
	require.NoError(t, err)
}

func TestSql2(t *testing.T) {
	t.Skip()
	DBStorage := NewDBStorage()
	timeout, cancelFunc := context.WithTimeout(context.Background(), time.Second*20)
	defer cancelFunc()
	err := DBStorage.Connect(timeout, DSN)
	require.NoError(t, err)

	err = DBStorage.DeleteEvent(timeout, "03cd6323-3590-45ec-a462-4e41dcffd8aa")
	require.NoError(t, err)
	err = DBStorage.DeleteEvent(timeout, "554d45c9-f8be-4de1-8152-3d4b88387055")
	require.NoError(t, err)
	err = DBStorage.Close()
	require.NoError(t, err)
}

func TestSql3(t *testing.T) {
	t.Skip()
	DBStorage := NewDBStorage()
	timeout, cancelFunc := context.WithTimeout(context.Background(), time.Second*20)
	defer cancelFunc()
	err := DBStorage.Connect(timeout, DSN)
	require.NoError(t, err)

	err = DBStorage.UpdateEvent(timeout, storage.Event{
		ID:          "03cd6323-3590-45ec-a462-4e41dcffd8aa",
		Title:       "some other title",
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(time.Second),
		Description: "some other description",
		OwnerID:     "123455678",
	})
	require.NoError(t, err)
	err = DBStorage.Close()
	require.NoError(t, err)
}

func TestFindSql(t *testing.T) {
	t.Skip()
	DBStorage := NewDBStorage()
	timeout, cancelFunc := context.WithTimeout(context.Background(), time.Second*20)
	defer cancelFunc()
	err := DBStorage.Connect(timeout, DSN)
	require.NoError(t, err)
	defer func() {
		err = DBStorage.Close()
		require.NoError(t, err)
	}()

	events, err := DBStorage.FindEventsInInterval(timeout, time.Now(), time.Now().AddDate(0, 0, 1))
	require.NoError(t, err)
	require.Equal(t, 2, len(events))
	fmt.Println(events)
}
