package whatsapp

import (
	"context"
	"fmt"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/sqlite"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// InitWaDB initializes the WhatsApp database connection
func InitWaDB(ctx context.Context, DBURI string) *sqlstore.Container {
	log = waLog.Stdout("Main", config.WhatsappLogLevel, true)
	dbLog := waLog.Stdout("Database", config.WhatsappLogLevel, true)

	storeContainer, err := initDatabase(ctx, dbLog, DBURI)
	if err != nil {
		log.Errorf("Database initialization error: %v", err)
		panic(pkgError.InternalServerError(fmt.Sprintf("Database initialization error: %v", err)))
	}

	return storeContainer
}

// initDatabase creates and returns a database store container based on the configured URI
func initDatabase(ctx context.Context, dbLog waLog.Logger, DBURI string) (*sqlstore.Container, error) {
	// Strip surrounding quotes that may come from .env file parsing
	DBURI = strings.TrimSpace(strings.Trim(DBURI, `"'`))
	if DBURI == "" {
		DBURI = config.DBURI
		if DBURI == "" {
			DBURI = "file:storages/whatsapp.db"
		}
	}

	if strings.HasPrefix(DBURI, "postgres:") || strings.HasPrefix(DBURI, "postgresql:") {
		return sqlstore.New(ctx, "postgres", DBURI, dbLog)
	}

	if !strings.HasPrefix(DBURI, "file:") {
		DBURI = "file:" + DBURI
	}
	DBURI = sqlite.FormatChatStorageURI(DBURI, true, true)
	return sqlstore.New(ctx, sqlite.DriverName, DBURI, dbLog)
}
