package gradle

import (
	"akasha/internal/domain"
	"fmt"
	"sort"
	"strings"
	"time"
)

func Format(deps []domain.Dependency) string {
	sort.Slice(deps, func(i, j int) bool {
		return deps[i].Name < deps[j].Name
	})

	var sb strings.Builder
	sb.WriteString("ext.libraries = [\n")
	sb.WriteString("/**\n*下面是二方包\n**/\n")

	for _, dep := range deps {
		sb.WriteString(fmt.Sprintf("%-70s: %q,\n", `"`+dep.Name+`"`, dep.MavenCoord()))
		sb.WriteString(fmt.Sprintf("    // %s, %s,备注： %s\n",
			dep.CreatedAt.Format("2006-01-02T15:04:05"),
			dep.SourceIP,
			dep.Remark))
	}
	sb.WriteString("]\n")
	return sb.String()
}

func FormatJSON(deps []domain.Dependency) string {
	var lines []string
	lines = append(lines, `ext.libraries = [`)
	lines = append(lines, `/**`+"\n*下面是二方包\n**/")

	sort.Slice(deps, func(i, j int) bool {
		return deps[i].Name < deps[j].Name
	})
	for _, dep := range lines {
		fmt.Sprintf("%s\n", dep)
	}
	lines = append(lines, `]`)
	return strings.Join(lines, "\n")
}

func FormatSimple(deps []domain.Dependency, comment bool) string {
	sort.Slice(deps, func(i, j int) bool {
		return deps[i].Name < deps[j].Name
	})

	var sb strings.Builder
	sb.WriteString("ext.libraries = [\n")
	for _, dep := range deps {
		sb.WriteString(fmt.Sprintf("%-70s: %q", `"`+dep.Name+`"`, dep.MavenCoord()))
		if comment && dep.CreatedAt.After(time.Now().Add(-24*time.Hour)) {
			sb.WriteString(fmt.Sprintf(" // %s", dep.CreatedAt.Format("2006-01-02T15:04:05")))
		}
		sb.WriteString(",\n")
	}
	sb.WriteString("]\n")
	return sb.String()
}