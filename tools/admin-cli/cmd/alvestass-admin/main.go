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

	buildClient := func() (*trailbase.Client, error) {
		tokens := &trailbase.Tokens{
			AuthToken:    cfg.AuthToken,
			RefreshToken: cfg.RefreshToken,
			CsrfToken:    cfg.CsrfToken,
		}
		c, err := trailbase.NewClientWithTokens(cfg.BackendURL, tokens)
		if err != nil {
			return nil, fmt.Errorf("anslutningsfel: %w", err)
		}
		return c, nil
	}

	client, err := buildClient()
	if err != nil {
		return err
	}

	// Startup connectivity check — also catches expired/invalid tokens.
	if _, err := client.GetClubInfo(); err != nil {
		return fmt.Errorf("anslutningsfel: bakänden är inte tillgänglig (%w)", err)
	}

	// Registered health checkers — append future checkers here.
	checkers := []validate.Checker{
		validate.NewClubInfoChecker(client),
	}

	handleOpErr := func(opErr error) error {
		if opErr == nil {
			return nil
		}
		if !trailbase.IsAuthError(opErr) {
			fmt.Fprintf(os.Stderr, "Fel: %v\n", opErr)
			return nil
		}
		fmt.Fprintln(os.Stderr, "Autentiseringsfel (403) — logga in igen.")
		updated, err := ui.RunSetup(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Inloggning misslyckades: %v\n", err)
			return nil
		}
		cfg = updated
		if err := cfg.Save(configPath); err != nil {
			fmt.Fprintf(os.Stderr, "Kunde inte spara konfiguration: %v\n", err)
		}
		newClient, err := buildClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Kunde inte skapa klient: %v\n", err)
			return nil
		}
		client = newClient
		checkers = []validate.Checker{validate.NewClubInfoChecker(client)}
		return nil
	}

	// Main menu loop.
	for {
		choice, err := ui.RunMenu()
		if err != nil {
			return err
		}
		switch choice {
		case ui.MenuUpdate:
			if err := handleOpErr(ui.RunUpdate(client)); err != nil {
				return err
			}
		case ui.MenuCheck:
			if err := handleOpErr(ui.RunCheck(checkers)); err != nil {
				return err
			}
		case ui.MenuHelp:
			ui.RunHelp()
		case ui.MenuImport:
			if err := handleOpErr(ui.RunImport(client)); err != nil {
				return err
			}
		case ui.MenuImportMembers:
			if err := handleOpErr(ui.RunImportMembers(client)); err != nil {
				return err
			}
		case ui.MenuQuit:
			fmt.Println("Hej då!")
			return nil
		}
	}
}
