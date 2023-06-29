package apimanager

import (
	"testing"

	"github.com/tidwall/gjson"
)

func TestGet(t *testing.T) {
	bs := []byte(`
	{
		"sn": "Hello",
		"step": "Hello",
		"tester": "Hello",
		"slot": "Hello",
		"userid": "611bc300-c4f5-4fe0-a2ca-41eb74572a81",
		"options": []
	  }
	`)

	r := gjson.GetBytes(bs, "options.#")
	t.Error(r.String())

}
