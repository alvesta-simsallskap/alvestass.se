package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/alvestass/admin-cli/internal/config"
	"github.com/alvestass/admin-cli/internal/trailbase"
	"github.com/alvestass/admin-cli/internal/ui"
	"github.com/alvestass/admin-cli/internal/validate"
)

var version = "dev"

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "sökväg till konfigurationsfil (standard: $UserConfigDir/alvestass-admin/config.json)")
	showVersion := flag.Bool("version", false, "visa version och avsluta")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Användning: alvestass-admin [flaggor]\n\nFlaggor:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Printf("alvestass-admin %s\n", version)
		os.Exit(0)
	}

	if err := run(configPath); err != nil {
		fmt.Fprintf(os.Stderr, "Fel: %v\n", err)
		os.Exit(1)
	}
}

func run(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("kunde inte läsa konfiguration: %w", err)
	}

	// First-run wizard (re-runs if credentials are missing).
	if !cfg.IsComplete() {
		updated, err := ui.RunSetup(cfg)
		if err != nil {
			return err
		}
		cfg = updated
		if err := cfg.Save(configPath); err != nil {
			return fmt.Errorf("kunde inte spara konfiguration: %w", err)
		}
	}

	tokens := &trailbase.Tokens{
		AuthToken:    cfg.AuthToken,
		RefreshToken: cfg.RefreshToken,
		CsrfToken:    cfg.CsrfToken,
	}
	client, err := trailbase.NewClientWithTokens(cfg.BackendURL, tokens)
	if err != nil {
		return fmt.Errorf("anslutningsfel: %w", err)
	}

	// Startup connectivity check — also catches expired/invalid tokens.
	if _, err := client.GetClubInfo(); err != nil {
		return fmt.Errorf("anslutningsfel: bakänden är inte tillgänglig (%w)", err)
	}

	// Registered health checkers — append future checkers here.
	checkers := []validate.Checker{
		validate.NewClubInfoChecker(client),
	}

	// Main menu loop.
	for {
		choice, err := ui.RunMenu()
		if err != nil {
			return err
		}
		switch choice {
		case ui.MenuUpdate:
			if err := ui.RunUpdate(client); err != nil {
				fmt.Fprintf(os.Stderr, "Fel: %v\n", err)
			}
		case ui.MenuCheck:
			if err := ui.RunCheck(checkers); err != nil {
				fmt.Fprintf(os.Stderr, "Fel: %v\n", err)
			}
		case ui.MenuHelp:
			ui.RunHelp()
		case ui.MenuImport:
			if err := ui.RunImport(client); err != nil {
				fmt.Fprintf(os.Stderr, "Fel: %v\n", err)
			}
		case ui.MenuQuit:
			fmt.Println("Hej då!")
			return nil
		}
	}
}
