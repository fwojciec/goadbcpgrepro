package main

import (
	"fmt"
	"os"

	"github.com/apache/arrow-adbc/go/adbc"
	"github.com/apache/arrow-adbc/go/adbc/drivermgr"
	"github.com/apache/arrow-go/v18/arrow/flight"
	"github.com/apache/arrow-go/v18/arrow/ipc"
)

func main() {
	if err := runServer(); err != nil {
		panic(err)
	}
}

type server struct {
	flight.BaseFlightServer
}

func (s *server) DoGet(tkt *flight.Ticket, fs flight.FlightService_DoGetServer) error {
	pgDsn := os.Getenv("POSTGRESQL_DSN")
	if pgDsn == "" {
		return fmt.Errorf("POSTGRESQL_DSN is not set")
	}

	var drv drivermgr.Driver
	db, err := drv.NewDatabase(map[string]string{
		"driver":          "adbc_driver_postgresql",
		adbc.OptionKeyURI: pgDsn,
	})
	if err != nil {
		return err
	}
	defer db.Close()
	conn, err := db.Open(fs.Context())
	if err != nil {
		return err
	}
	defer conn.Close()

	st, err := conn.NewStatement()
	if err != nil {
		return err
	}
	defer st.Close()

	if err := st.SetSqlQuery("SELECT * FROM synthetic_filter_option"); err != nil {
		return err
	}

	reader, _, err := st.ExecuteQuery(fs.Context())
	if err != nil {
		return err
	}
	defer reader.Release()

	wr := flight.NewRecordWriter(
		fs,
		ipc.WithSchema(reader.Schema()),
	)
	defer wr.Close()

	for reader.Next() {
		rec := reader.Record()
		if err := wr.Write(rec); err != nil {
			return fmt.Errorf("couldn't write record: %w", err)
		}
	}

	return nil
}

func runServer() error {
	s := flight.NewServerWithMiddleware(nil)
	s.Init("0.0.0.0:8816")
	s.RegisterFlightService(&server{})

	return s.Serve()
}
