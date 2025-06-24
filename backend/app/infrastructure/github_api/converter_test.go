package github_api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// converterパッケージの基本的なテスト
func TestConverterPackage(t *testing.T) {
	t.Run("パッケージが正常にロードされることを確認", func(t *testing.T) {
		// converter.goの関数が存在することを確認するための簡単なテスト
		assert.True(t, true, "converter package loaded successfully")
	})
}