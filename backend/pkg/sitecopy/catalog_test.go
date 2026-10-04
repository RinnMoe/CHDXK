package sitecopy

import (
	"context"
	"testing"
)

func TestCopyParametersAndContext(t *testing.T) {
	var key string
	for k, e := range Entries() {
		if len(Parameters(e.Text)) > 0 {
			key = k
			break
		}
	}
	if key == "" {
		t.Fatal("missing parameterized entries")
	}
	original := Entries()[key].Text
	if !ValidValue(key, original) {
		t.Fatal("default rejected")
	}
	if ValidValue(key, "没有内容标记") || ValidValue(key, original+"{v999}") {
		t.Fatal("missing or extra parameter accepted")
	}
	if ValidValue(key, "") {
		t.Fatal("empty required text accepted")
	}
	if !ValidValue("copy.brand.feedback_email", "") {
		t.Fatal("optional email must allow clearing")
	}
	var serverKey, serverText string
	for k, e := range Entries() {
		if len(k) > 12 && k[:12] == "copy.server." {
			serverKey = k
			serverText = e.Text
			break
		}
	}
	ctx := WithValues(context.Background(), map[string]string{serverKey: "自定义错误说明"})
	if Text(ctx, serverText) != "自定义错误说明" || Text(ctx, "unknown") != "unknown" {
		t.Fatal("context translation failed")
	}
	if Text(context.Background(), serverText) != serverText {
		t.Fatal("override leaked to another request")
	}
}
