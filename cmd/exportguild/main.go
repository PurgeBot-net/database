package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/PurgeBot-net/database"
)

var (
	guildID = flag.Int64("guildid", 0, "guild ID to export")
	db      = flag.String("db", "service=purgebot-local", `libpq connection string; production is "service=purgebot-prod". Services live in %APPDATA%\postgresql\.pg_service.conf, passwords in %APPDATA%\postgresql\pgpass.conf`)
)

func main() {
	flag.Parse()
	if *guildID <= 0 || flag.NArg() > 0 {
		flag.Usage()
		os.Exit(2)
	}
	ctx := context.Background()

	fmt.Printf("Exporting guild data for %d\n", *guildID)
	cfg, err := pgconn.ParseConfig(*db)
	must(err, "read -db")
	target := fmt.Sprintf("%s@%s:%d/%s", cfg.User, cfg.Host, cfg.Port, cfg.Database)
	fmt.Printf("[1/3] Connecting to %s...\n", target)
	conn, err := database.New(ctx, *db)
	must(err, "connect")
	defer conn.Close()

	fmt.Println("[2/3] Reading rows...")
	e, err := conn.ExportGuild(ctx, *guildID)
	must(err, "export")

	fmt.Println("[3/3] Writing zip...")
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	total := 0
	for _, name := range slices.Sorted(maps.Keys(e.Tables)) {
		var rows []json.RawMessage
		must(json.Unmarshal(e.Tables[name], &rows), "read "+name)
		var pretty bytes.Buffer
		must(json.Indent(&pretty, e.Tables[name], "", "  "), "format "+name)
		addFile(zw, name+".json", pretty.Bytes())
		fmt.Printf("  - %s.json: %d row(s)\n", name, len(rows))
		total += len(rows)
	}
	if total == 0 {
		fmt.Printf("Nothing stored for guild %d on %s; no zip written. Check the ID and -db.\n", *guildID, target)
		return
	}
	if len(e.Avatar) > 0 {
		ext, ok := strings.CutPrefix(http.DetectContentType(e.Avatar), "image/")
		if !ok {
			ext = "bin"
		}
		addFile(zw, "bot_avatar."+ext, e.Avatar)
		fmt.Printf("  - bot_avatar.%s: %d KB\n", ext, (len(e.Avatar)+1023)/1024)
	}
	must(zw.Close(), "zip")

	dir := "exports"
	must(os.MkdirAll(dir, 0o700), "create "+dir)
	path := filepath.Join(dir, fmt.Sprintf("guild-%d.zip", *guildID))
	must(os.WriteFile(path, buf.Bytes(), 0o600), "write "+path)
	fmt.Printf("Done! Export saved to %s; delete it once it has been delivered.\n", path)
}

func addFile(zw *zip.Writer, name string, data []byte) {
	w, err := zw.Create(name)
	must(err, "zip "+name)
	_, err = w.Write(data)
	must(err, "zip "+name)
}

func must(err error, what string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s: %v\n", what, err)
		os.Exit(1)
	}
}
