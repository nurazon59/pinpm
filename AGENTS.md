# AGENTS.md

## テスト

- **Table-driven tests**: 複数のケースがあるテストは必ずtable-drivenで書く
- **assertライブラリ**: `github.com/stretchr/testify` の `assert`/`require` を使用する
  - `require.NoError` → エラー時に即fail
  - `assert.Equal` → 値の比較
  - `assert.Error` → エラー期待値
- `t.Errorf`/`t.Fatalf` は使用しない
