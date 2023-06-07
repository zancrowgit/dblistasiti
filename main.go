package main

import (
	"fmt"
    "log"
    "os"
	"net"
	"time"
	"strconv"
	"strings"
	
	"github.com/zancrowgit/dbpostgres"
	_ "github.com/lib/pq"
    "github.com/jmoiron/sqlx" 
	_ "github.com/godror/godror"
	"github.com/tatsushid/go-fastping"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var dbRAM *sqlx.DB  // RAM ORACLE
var err   error

const (
    DB_USER      = "reqass"
    DB_PASSWORD  = "reqass"
    DB_NAME      = "req"
)   


type VPN struct {
	ID_VPN       int    `db:"ID_VPN"          json:"ID"` 
	COD_COMMESSA string `db:"COD_COMMESSA"    json:"COD_COMMESSA"`
	CLASSE_VPN   string `db:"CLASSE_VPN"      json:"CLASSE_VPN"`
	IP_PUBBLICO	 string `db:"IP_PUBBLICO_VPN" json:"IP_PUBBLICO"`
	DESCRIZIONE  string `db:"DESCRIZIONE"     json:"DESCRIZIONE"`
	PING		 string `db:"PING"			  json:"PING"`
}


// LOGGER ZAP
func fileLogger(filename string) *zap.Logger {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(config)
	logFile, _ := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	writer := zapcore.AddSync(logFile)
	defaultLogLevel := zapcore.DebugLevel
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, defaultLogLevel),
	)
  
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
  
	return logger
 }



func main(){
	////////////////////////////////////////////////////////////////
	// SET LOGGER
	filename := "logs.log"
	logger := fileLogger(filename)
  
	//logger.Info("INFO log level message")
	//logger.Warn("Warn log level message")
	//logger.Error("Error log level message")


	/////////////////////////////////////////////////////////////////
	// REQ RAM
	// connetti DB ram
	dbRAM, err = sqlx.Connect("godror", DB_USER+"/"+DB_PASSWORD+"@172.16.8.114:1521/req")
	if err != nil {
		log.Fatalln(err)
	}
	/////////////////////////////////////////////////////////////////
	// POSTGRES 
	//
	dbpostgres := database.ConnectDB()

	for {	// LOOP INFINITO
		
		logger.Info("inizio task")


		/////////////////////////////////////////////////////////////////
		// GET VPNs E PING ---> Assegna a Struct
		//
		VPNs := getVPN()     // Prendi info dal DB di Ram e assegna a una Struct di tipo []VPN
		
		newVPN := []VPN{}    // Assegna ad una struct sempre di tipo VPN   VUOTA per appendere nuovo contenuto con ping.
		for _, vpn := range VPNs {	  
			vpn.PING = getPing(vpn.IP_PUBBLICO)    // PING VERSO IP PUBBLICO
			newVPN = append(newVPN, vpn)	       // Travasa contenuto vecchia VPNs però col Ping nuovo.
		} 


		/////////////////////////////////////////////////////////////////
		// POSTGRES 
		//
		
		dbpostgres.Exec("DELETE FROM vpns")     // 1° Troncate Table
		result := dbpostgres.Create(&newVPN)    // 2° INSERT ALL DATA
		if result.Error != nil {
			logger.Error("Errore Insert Riga 133.")
		}

		var count int64
		dbpostgres.Table("vpns").Count(&count)
		stringa := strconv.FormatInt(count, 10) // s == "97" (decimal)
		logger.Info("Count Tabella : ")
		logger.Info(stringa)
		logger.Info("fine task")


		time.Sleep(3600 * time.Second)
		
	}
	defer dbRAM.Close()  // chiudi DB ram
		

}

// da mettere logs
func getPing (host string) string {     // FAST PING --> Ping ICMP
    var ping bool
    ping = false
    p := fastping.NewPinger()

    p.AddIP(strings.Trim(host, " "))
    p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		//fmt.Printf("IP Addr: %s receive, RTT: %v\n", addr.String(), rtt)
        ping = true
    }

    err = p.Run()
    if err != nil {
        fmt.Println(err)
    } 


    if ping {
        return "1"
    }else{
        return "0"
    }
}

// DA METTERE LOGS
func getVPN() []VPN{
	VPNss := []VPN{}
	rows, err := dbRAM.Queryx("SELECT v.ID_VPN, v.COD_COMMESSA, v.CLASSE_VPN, v.IP_PUBBLICO_VPN, c.DESCRIZIONE  FROM VPN v INNER JOIN COMMESSE c ON v.COD_COMMESSA = c.COD_COMMESSA ORDER BY v.ID_VPN")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var record VPN
		if err := rows.StructScan(&record); err != nil {
			panic(err)
		}
        VPNss = append(VPNss, record)
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}
	return VPNss
} 





