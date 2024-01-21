package elog

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/clarktrimble/delish/elog/logmsg"
)

func TestMinLog(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MinLog Suite")
}

var _ = Describe("MinLog", func() {

	var (
		ctx context.Context
		lgr *MinLog
	)

	Describe("stashing fields in the ctx", func() {
		var (
			wfCtx context.Context
			//anotherCtx context.Context
			kvs []any
		)

		BeforeEach(func() {
			ctx = context.Background()
			lgr = &MinLog{}
		})

		JustBeforeEach(func() {
			wfCtx = lgr.WithFields(ctx, kvs...)
		})

		When("a string field is added to bg ctx", func() {
			BeforeEach(func() {
				kvs = []any{"foo", "bar"}
			})

			It("should store it as a field in the returned ctx", func() {
				Expect(logmsg.GetFields(wfCtx)).To(Equal(logmsg.Fields{
					"foo": {Data: []byte("\"bar\""), Quoted: true},
				}))
			})
		})

		When("adding a variety of numbers to a ctx", func() {
			BeforeEach(func() {
				kvs = []any{
					"an_integer", 34534,
					"a_negative_integer", -888,
					"a_large_integer", uint64(1<<64 - 1),
					"a_float", 2.71828,
				}
			})

			It("should pop them into ctx", func() {
				Expect(logmsg.GetFields(wfCtx)).To(Equal(logmsg.Fields{
					"an_integer":         {Data: []byte("34534")},
					"a_negative_integer": {Data: []byte("-888")},
					"a_large_integer":    {Data: []byte("18446744073709551615")},
					"a_float":            {Data: []byte("2.71828")},
				}))
			})
		})

		When("adding time, duration, and bool to a ctx", func() {
			BeforeEach(func() {
				kvs = []any{
					"a_time", time.Time{},
					"a_duration", time.Minute,
					"a_bool", true,
				}
			})

			It("should pop them into ctx", func() {
				Expect(logmsg.GetFields(wfCtx)).To(Equal(logmsg.Fields{
					"a_time":     {Data: []byte(`"0001-01-01T00:00:00Z"`), Quoted: true},
					"a_duration": {Data: []byte("60000000000")},
					"a_bool":     {Data: []byte("true")},
				}))
			})
		})

		When("adding stuff to a ctx", func() {
			BeforeEach(func() {
				kvs = []any{
					"a_map", map[string]any{"one": 55},
					"a_slice", []string{"one", "two"},
					"a_null_value", logmsg.Value{},
					"an_obj", demo{One: "one", Two: 2},
				}
			})

			It("should stringify and pop them into ctx", func() {
				Expect(logmsg.GetFields(wfCtx)).To(Equal(logmsg.Fields{
					"a_map":        {Data: []byte(`"{"one":55}"`), Quoted: true},
					"a_slice":      {Data: []byte(`"["one","two"]"`), Quoted: true},
					"a_null_value": {Data: []byte("null")},
					"an_obj":       {Data: []byte(`"{"One":"one","Two":2}"`), Quoted: true},
				}))
			})
		})

		/*

			When("adding fields to the same ctx", func() {
				BeforeEach(func() {
					ctx = lgr.WithFields(ctx, "ta", "dum")
				})

				It("should accumulate args as fields in ctx", func() {
					Expect(getFields(wfCtx)).To(Equal(fields{
						"ta":  "dum",
						"foo": "bar",
					}))
				})
			})

			When("adding fields with same key to the same ctx", func() {
				BeforeEach(func() {
					ctx = lgr.WithFields(ctx, "foo", "baz")
				})

				It("should overwrite that key", func() {
					Expect(getFields(wfCtx)).To(Equal(fields{
						"foo": "bar",
					}))
				})
			})

			When("adding fields to a different ctx", func() {
				BeforeEach(func() {
					ctx = lgr.WithFields(ctx, "som", "mat")
					anotherCtx = lgr.WithFields(ctx, "gar", "ble")
				})

				It("should not cross pollinate, thanks to copyFields", func() {
					Expect(getFields(wfCtx)).To(Equal(fields{
						"som": "mat",
						"foo": "bar",
					}))
					Expect(getFields(anotherCtx)).To(Equal(fields{
						"som": "mat",
						"gar": "ble",
					}))
				})
			})

			When("adding a nice variety of fields to a ctx", func() {
				BeforeEach(func() {
					kvs = []any{
						"a_string", "This is a string field value.",
						"a_bool", true,
						"a_time", time.Time{},
						"a_duration", time.Minute,
						"an_integer", 34534,
						"a_negative_integer", -888,
						"a_large_integer", uint64(1<<64 - 1),
						"a_float", 2.71828,
						"a_map", map[string]any{"one": 55},
						"a_slice", []string{"one", "two"},
						"an_obj", &MinLog{},
					}
				})

				It("should stringify and pop them into ctx", func() {
					Expect(getFields(wfCtx)).To(Equal(fields{
						"a_string":           "This is a string field value.",
						"a_bool":             "true",
						"a_time":             "0001-01-01 00:00:00 +0000 UTC", // Todo: tweak??
						"a_duration":         "1m0s",                          // Todo: tweak??
						"an_integer":         "34534",
						"a_large_integer":    "18446744073709551615",
						"a_negative_integer": "-888",
						"a_float":            "2.71828",
						"a_map":              "{\"one\":55}", // escaped! a good thing yeah
						"a_slice":            "[\"one\",\"two\"]",
						"an_obj":             "{\"Writer\":null,\"AltWriter\":null,\"Formatter\":null,\"MaxLen\":0}",
					}))
				})
			})

			When("adding an odd number of, um, kv's to a ctx", func() {
				BeforeEach(func() {
					kvs = []any{
						"a_string", "This is a string field value.",
						"odd man out here",
					}
				})

				It("should reformulate into a helpful logerror", func() {
					Expect(getFields(wfCtx)).To(Equal(fields{
						"a_string": "This is a string field value.",
						"logerror": "no field name found for: odd man out here",
					}))
				})
			})

			When("adding garbage to a ctx", func() {
				BeforeEach(func() {
					kvs = []any{
						"a_string", "This is a string field value.",
						"a_channel", make(chan int),
					}
				})

				It("should reformulate into a helpful logerror", func() {
					flds := getFields(wfCtx)
					Expect(flds).To(HaveKey("a_string"))
					Expect(flds["a_string"]).To(Equal("This is a string field value."))
					Expect(flds).To(HaveKey("logerror"))
					Expect(flds["logerror"]).To(HavePrefix("failed to marshal value: (chan int)"))
				})
			})
		*/

	})
})

type demo struct {
	One string
	Two int
}
