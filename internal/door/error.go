package door

import (
	"fmt"

	"github.com/doors-dev/gox"
)

type renderError struct {
	err error
	id  uint64
}

func (e renderError) Main() gox.Elem {

	return gox.Elem(func(cur gox.Cursor) error {
		if err := cur.Init("div"); err != nil {
			return err
		}
		{
			if err := cur.AttrSet("role", "alert"); err != nil {
				return err
			}
			if err := cur.AttrSet("aria-live", "polite"); err != nil {
				return err
			}
			if err := cur.AttrSet("data-fw", "error"); err != nil {
				return err
			}
			if err := cur.Submit(); err != nil {
				return err
			}
			if err := cur.Text(`Component Error. `); err != nil {
				return err
			}
			if err := cur.Init("span"); err != nil {
				return err
			}
			{
				if err := cur.AttrSet("data-fw", "error-id"); err != nil {
					return err
				}
				if err := cur.Submit(); err != nil {
					return err
				}
				if err := cur.Text(fmt.Sprintf(`ID: %d`, e.id)); err != nil {
					return err
				}
			}
			if err := cur.Close(); err != nil {
				return err
			}
			if err := cur.Raw(fmt.Sprintf(`<!-- %s -->`, e.err.Error())); err != nil {
				return err
			}
		}
		if err := cur.Close(); err != nil {
			return err
		}
		return nil
	})
}
