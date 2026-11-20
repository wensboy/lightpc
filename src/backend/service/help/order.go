package serviceHelpImpl

import (
	"bytes"
	"encoding/json"
	"fmt"

	serviceHelp "github.com/wensboy/lightpc/backend/domain/service/help"
)

// tocGroups decodes the toc object while keeping the declaration order, because the tree shows the groups
// in the order the author wrote them and a map would lose it.
type tocGroups []serviceHelp.DocsGroup

// UnmarshalJSON walks the object token by token so the order survives.
func (t *tocGroups) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '{' {
		return fmt.Errorf("toc must be an object")
	}

	groups := make(tocGroups, 0, 8)
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return err
		}
		title, _ := key.(string)
		var items tocItems
		if err := decoder.Decode(&items); err != nil {
			return err
		}
		groups = append(groups, serviceHelp.DocsGroup{Title: title, Items: items})
	}
	*t = groups
	return nil
}

// tocItems decodes one group of documents while keeping the declaration order.
type tocItems []serviceHelp.DocsItem

// UnmarshalJSON walks the object token by token so the order survives.
func (i *tocItems) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if _, err := decoder.Token(); err != nil {
		return err
	}

	items := make(tocItems, 0, 8)
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return err
		}
		title, _ := key.(string)
		var target string
		if err := decoder.Decode(&target); err != nil {
			return err
		}
		items = append(items, serviceHelp.DocsItem{Title: title, Path: target})
	}
	*i = items
	return nil
}
