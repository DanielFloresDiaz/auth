package templatemailer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"github.com/supabase/auth/internal/conf"
)

func TestProjectTemplateSlugLowercasesName(t *testing.T) {
	require.Equal(t, "zion", projectTemplateSlug("ZION"))
	require.Equal(t, "iec", projectTemplateSlug("IEC"))
	require.Equal(t, "brawler", projectTemplateSlug("BRAWLER"))
}

func TestProjectTemplatePathRejectsTraversal(t *testing.T) {
	_, ok := projectTemplatePath("/templates", "..", projectThemeFileName)
	require.False(t, ok)
}

func TestThemeForProjectUsesProjectDirThemeYAML(t *testing.T) {
	projectID := uuid.Must(uuid.NewV4())
	base := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(base, "zion"), 0o755))
	themeYAML := `brand:
  product_name: Custom Zion
colors:
  accent: "#E6A815"
`
	require.NoError(t, os.WriteFile(filepath.Join(base, "zion", projectThemeFileName), []byte(themeYAML), 0o644))

	cfg := &conf.GlobalConfiguration{}
	cfg.Mailer.Templates.ProjectDir = base

	cache := NewCache()
	cache.ProjectNameLookup = func(_ context.Context, id uuid.UUID) (string, error) {
		require.Equal(t, projectID, id)
		return "ZION", nil
	}

	theme := cache.themeForProject(context.Background(), cfg, projectID.String())
	require.Equal(t, "Custom Zion", theme.Brand.ProductName)
	require.Equal(t, "#E6A815", theme.Colors.Accent)
	require.Equal(t, "#eef0f3", theme.Colors.PageBg)
}

func TestThemeForProjectMissingFileReturnsDefault(t *testing.T) {
	projectID := uuid.Must(uuid.NewV4())
	base := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(base, "brawler"), 0o755))

	cfg := &conf.GlobalConfiguration{}
	cfg.Mailer.Templates.ProjectDir = base

	cache := NewCache()
	cache.ProjectNameLookup = func(_ context.Context, _ uuid.UUID) (string, error) {
		return "BRAWLER", nil
	}

	theme := cache.themeForProject(context.Background(), cfg, projectID.String())
	require.Equal(t, DefaultProjectTheme().Colors.PageBg, theme.Colors.PageBg)
	require.Empty(t, theme.Brand.ProductName)
}

func TestThemeForProjectInvalidYAMLReturnsDefault(t *testing.T) {
	projectID := uuid.Must(uuid.NewV4())
	base := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(base, "iec"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(base, "iec", projectThemeFileName), []byte(":\n- bad"), 0o644))

	cfg := &conf.GlobalConfiguration{}
	cfg.Mailer.Templates.ProjectDir = base

	cache := NewCache()
	cache.ProjectNameLookup = func(_ context.Context, _ uuid.UUID) (string, error) {
		return "IEC", nil
	}

	theme := cache.themeForProject(context.Background(), cfg, projectID.String())
	require.Equal(t, DefaultProjectTheme().Copy.Invite.Title, theme.Copy.Invite.Title)
}

func TestMergeProjectThemePreservesDefaults(t *testing.T) {
	base := DefaultProjectTheme()
	merged := mergeProjectTheme(base, ProjectTheme{
		Brand:  ProjectThemeBrand{ProductName: "Brawler"},
		Colors: ProjectThemeColors{Accent: "#ff8000"},
	})
	require.Equal(t, "Brawler", merged.Brand.ProductName)
	require.Equal(t, "#ff8000", merged.Colors.Accent)
	require.Equal(t, base.Copy.Invite.CtaLabel, merged.Copy.Invite.CtaLabel)
}
