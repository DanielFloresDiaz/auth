package templatemailer

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
)

const projectThemeFileName = "theme.yaml"

// ProjectNameLookup resolves a project display name from its id (for template paths).
type ProjectNameLookup func(ctx context.Context, projectID uuid.UUID) (string, error)

func projectTemplateSlug(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func projectTemplatePath(projectDir, slug, fileName string) (string, bool) {
	cleanDir := filepath.Clean(projectDir)
	path := filepath.Join(cleanDir, slug, fileName)
	if !strings.HasPrefix(path, cleanDir+string(filepath.Separator)) {
		return "", false
	}
	return path, true
}

func invalidProjectID(projectID string) bool {
	return projectID == "" || projectID == "00000000-0000-0000-0000-000000000000"
}

func (o *Cache) projectSlug(ctx context.Context, projectID string) string {
	if o.ProjectNameLookup == nil || invalidProjectID(projectID) {
		return ""
	}

	o.rw.RLock()
	if name, ok := o.projectNames[projectID]; ok {
		o.rw.RUnlock()
		return projectTemplateSlug(name)
	}
	o.rw.RUnlock()

	id, err := uuid.FromString(projectID)
	if err != nil {
		return ""
	}

	name, err := o.ProjectNameLookup(ctx, id)
	if err != nil {
		logrus.WithError(err).WithField("project_id", projectID).Warn("project template: name lookup failed; per-project theme skipped")
		return ""
	}
	if strings.TrimSpace(name) == "" {
		return ""
	}

	o.rw.Lock()
	if o.projectNames == nil {
		o.projectNames = make(map[string]string)
	}
	o.projectNames[projectID] = name
	o.rw.Unlock()

	return projectTemplateSlug(name)
}
