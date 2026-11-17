package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/PurgeBot-net/database"
)

var (
	guildID = flag.Int64("guildid", 0, "guild ID to delete")
	db      = flag.String("db", "service=purgebot-local", `libpq connection string; production is "service=purgebot-prod". Services live in %APPDATA%\postgresql\.pg_service.conf, passwords in %APPDATA%\postgresql\pgpass.conf`)
	confirm = flag.Bool("confirm", false, "delete the rows; without it the run is rolled back")
)

func main() {
	flag.Parse()
	if *guildID <= 0 || flag.NArg() > 0 {
		flag.Usage()
		os.Exit(2)
	}
	ctx := context.Background()

	fmt.Printf("Deleting guild data for %d\n", *guildID)
	cfg, err := pgconn.ParseConfig(*db)
	must(err, "read -db")
	target := fmt.Sprintf("%s@%s:%d/%s", cfg.User, cfg.Host, cfg.Port, cfg.Database)
	fmt.Printf("[1/2] Connecting to %s...\n", target)
	conn, err := database.New(ctx, *db)
	must(err, "connect")
	defer conn.Close()

	if *confirm {
		fmt.Println("[2/2] Deleting rows...")
	} else {
		fmt.Println("[2/2] Dry run...")
	}
	rows, err := conn.DeleteGuild(ctx, database.DeleteGuildParams{GuildID: *guildID, Confirm: *confirm})
	must(err, "delete")

	var total int64
	for _, r := range rows {
		total += r.Deleted + r.Kept
	}
	if total == 0 {
		fmt.Printf("  - Nothing stored in Postgres for guild %d on %s; check the ID and -db. Redis is not checked.\n", *guildID, target)
	}
	for _, r := range rows {
		if r.Kept > 0 {
			fmt.Printf("  - %s: %d deleted, %d kept (stats opt-out)\n", r.Table, r.Deleted, r.Kept)
		} else {
			fmt.Printf("  - %s: %d deleted\n", r.Table, r.Deleted)
		}
	}
	fmt.Println("  - guild_departures: set to 1970; the next sweep deletes what is kept and the Redis job keys if the bot has left, or keeps the opt-out if it is still there")

	if !*confirm {
		fmt.Println("Dry run, rolled back. Run again with -confirm to delete.")
		return
	}
	fmt.Println("Deleted.")
}

func must(err error, what string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s: %v\n", what, err)
		os.Exit(1)
	}
}
