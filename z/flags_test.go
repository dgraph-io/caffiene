package z

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFlag(t *testing.T) {
	const opt = `bool_key=true; int-key=5; float-key=0.05; string_key=value; ;`
	const def = `bool_key=false; int-key=0; float-key=1.0; string-key=; other-key=5;
		duration-minutes=15m; duration-hours=12h; duration-days=30d;`

	_, err := NewSuperFlag("boolo-key=true").mergeAndCheckDefaultImpl(def)
	require.Error(t, err)
	_, err = newSuperFlagImpl("key-without-value")
	require.Error(t, err)

	// bool-key and int-key should not be overwritten. Only other-key should be set.
	sf := NewSuperFlag(opt)
	sf.MergeAndCheckDefault(def)

	require.Equal(t, true, sf.GetBool("bool-key"))
	require.Equal(t, uint64(5), sf.GetUint64("int-key"))
	require.Equal(t, "value", sf.GetString("string-key"))
	require.Equal(t, uint64(5), sf.GetUint64("other-key"))

	require.Equal(t, time.Minute*15, sf.GetDuration("duration-minutes"))
	require.Equal(t, time.Hour*12, sf.GetDuration("duration-hours"))
	require.Equal(t, time.Hour*24*30, sf.GetDuration("duration-days"))
}

func TestFlagDefault(t *testing.T) {
	def := `one=false; two=; three=;`
	f := NewSuperFlag(`one=true; two=4;`).MergeAndCheckDefault(def)
	require.Equal(t, true, f.GetBool("one"))
	require.Equal(t, int64(4), f.GetInt64("two"))
}
