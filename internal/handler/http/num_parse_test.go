package http

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNumParse(t *testing.T) {
	raw := `{ "a": 1209600 }`

	m := map[string]any{}

	err := json.Unmarshal([]byte(raw), &m)
	require.NoError(t, err)

	fmt.Println(m)

	v := m["a"]

	str := fmt.Sprintf("%v", v)

	fmt.Println(str)

	num2, err := strconv.ParseFloat(str, 64)
	require.NoError(t, err)

	fmt.Println(int64(num2))
}
