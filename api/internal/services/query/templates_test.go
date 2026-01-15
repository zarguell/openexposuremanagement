package query

import (
	"testing"
)

func TestQueryTemplates(t *testing.T) {
	t.Run("load missing software template", func(t *testing.T) {
		tmpl := GetTemplate("missing_software")
		if tmpl == nil {
			t.Fatal("expected template, got nil")
		}

		if tmpl.Name != "Assets Missing Critical Software" {
			t.Errorf("got name %s, want 'Assets Missing Critical Software'", tmpl.Name)
		}

		if tmpl.PrimaryEntity != "assets" {
			t.Errorf("got primary_entity %s, want 'assets'", tmpl.PrimaryEntity)
		}

		if tmpl.Join == nil {
			t.Error("expected join in missing software template")
		}

		if tmpl.Join.Entity != "software_inventory" {
			t.Errorf("got join entity %s, want 'software_inventory'", tmpl.Join.Entity)
		}
	})

	t.Run("load exploitable CVEs template", func(t *testing.T) {
		tmpl := GetTemplate("exploitable_cves")
		if tmpl == nil {
			t.Fatal("expected template, got nil")
		}

		if tmpl.PrimaryEntity != "assets" {
			t.Errorf("got primary_entity %s, want 'assets'", tmpl.PrimaryEntity)
		}

		if tmpl.Join == nil {
			t.Error("expected join in exploitable CVEs template")
		}

		if tmpl.Join.Entity != "findings" {
			t.Errorf("got join entity %s, want 'findings'", tmpl.Join.Entity)
		}
	})

	t.Run("load software vulnerabilities template", func(t *testing.T) {
		tmpl := GetTemplate("software_vulnerabilities")
		if tmpl == nil {
			t.Fatal("expected template, got nil")
		}

		if tmpl.PrimaryEntity != "findings" {
			t.Errorf("got primary_entity %s, want 'findings'", tmpl.PrimaryEntity)
		}
	})

	t.Run("list all templates", func(t *testing.T) {
		templates := ListTemplates()
		if len(templates) != 3 {
			t.Errorf("expected 3 templates, got %d", len(templates))
		}
	})
}
