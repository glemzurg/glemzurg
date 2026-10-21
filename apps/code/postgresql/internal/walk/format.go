package walk

import (
	"fmt"
	"strings"
)

// Format renders a walk as stable text for the command output.
func Format(result Result) string {
	var b strings.Builder
	fmt.Fprintf(&b, "model: %s%s\n", result.ModelName, includeSuffix(result.Include))
	fmt.Fprintf(&b, "key: %s\n", result.ModelKey)
	if result.Schema != "" {
		fmt.Fprintf(&b, "schema: %s\n", result.Schema)
	}
	writeAssocs(&b, result.Associations)
	for _, domain := range result.Domains {
		fmt.Fprintf(&b, "domain: %s name=%s%s\n", domain.Key, domain.Name, includeSuffix(domain.Include))
		writeAssocs(&b, domain.Associations)
		for _, sub := range domain.Subdomains {
			writeSubdomain(&b, domain.Key, sub)
		}
	}
	return b.String()
}

func writeSubdomain(b *strings.Builder, domainKey string, sub SubdomainVisit) {
	fmt.Fprintf(b, "subdomain: %s/%s name=%s%s\n", domainKey, sub.Key, sub.Name, includeSuffix(sub.Include))
	for _, class := range sub.Classes {
		label := domainKey + "/" + sub.Key + "/" + class.Key
		fmt.Fprintf(b, "class: %s name=%s attributes=%d%s\n", label, class.Name, class.Attributes, includeSuffix(class.Include))
		for _, attr := range class.Noted {
			writeAttribute(b, label, attr)
		}
	}
	writeAssocs(b, sub.Associations)
}

func writeAttribute(b *strings.Builder, classLabel string, attr AttrVisit) {
	fmt.Fprintf(b, "attribute: %s.%s", classLabel, attr.Key)
	if attr.Type != "" {
		fmt.Fprintf(b, " type=%s", attr.Type)
	}
	if attr.Default != "" {
		fmt.Fprintf(b, " default=%s", attr.Default)
	}
	b.WriteString(includeSuffix(attr.Include))
	b.WriteByte('\n')
}

func writeAssocs(b *strings.Builder, assocs []AssocVisit) {
	for _, assoc := range assocs {
		fmt.Fprintf(b, "association: %s", assoc.Label)
		if assoc.OnDelete != "" {
			fmt.Fprintf(b, " on_delete=%s", assoc.OnDelete)
		}
		b.WriteString(includeSuffix(assoc.Include))
		b.WriteByte('\n')
	}
}

func includeSuffix(value *bool) string {
	if value == nil {
		return ""
	}
	if *value {
		return " include=true"
	}
	return " include=false"
}
