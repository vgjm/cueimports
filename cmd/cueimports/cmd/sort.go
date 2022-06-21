package cmd

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Official < Common(Third Party) < Same module
const (
	OfficialWeight = iota - 1
	CommonWeight
	SameModWeight
)

type Entry struct {
	Contents string
	Weight   int
}

type List struct {
	Module     string
	RawEntries []string
	Entries    []Entry
}

func (l List) Len() int {
	return len(l.Entries)
}

func (l List) Less(i, j int) bool {
	if l.Entries[i].Weight != l.Entries[j].Weight {
		return l.Entries[i].Weight < l.Entries[j].Weight
	}
	return l.Entries[i].Contents < l.Entries[j].Contents
}

func (l List) Swap(i, j int) {
	l.Entries[i], l.Entries[j] = l.Entries[j], l.Entries[i]
	l.RawEntries[i], l.RawEntries[j] = l.RawEntries[j], l.RawEntries[i]
}

func sortEntries(entries []string, module string) error {
	// Tell if the entry belongs to the module.
	modRe, err := regexp.Compile("^" + module)
	if err != nil {
		return fmt.Errorf("failed to compile regular expression for module: %s", module)
	}
	l := List{
		Module:     module,
		RawEntries: entries,
		Entries:    make([]Entry, len(entries)),
	}
	// Find the real entry instead of the one with alias.
	EntryRe := regexp.MustCompile(`\"(.*)\"`)
	for i := range l.RawEntries {
		res := EntryRe.FindStringSubmatch(l.RawEntries[i])
		if len(res) < 2 {
			return fmt.Errorf("can't match entry: %s", l.RawEntries[i])
		}
		l.Entries[i].Contents = res[1]
		if !strings.Contains(l.Entries[i].Contents, ".") {
			l.Entries[i].Weight = OfficialWeight
		}
		if module != "" && modRe.MatchString(l.Entries[i].Contents) {
			l.Entries[i].Weight = SameModWeight
		}
	}
	sort.Sort(l)
	return nil
}
