package json

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"Dana"
	"Dana/internal/fuzz"
	"Dana/metric"
	"Dana/testutil"
)

const (
	validJSON              = "{\"a\": 5, \"b\": {\"c\": 6}}"
	validJSONNewline       = "\n{\"d\": 7, \"b\": {\"d\": 8}}\n"
	validJSONArray         = "[{\"a\": 5, \"b\": {\"c\": 6}}]"
	validJSONArrayMultiple = "[{\"a\": 5, \"b\": {\"c\": 6}}, {\"a\": 7, \"b\": {\"c\": 8}}]"
	invalidJSON            = "I don't think this is JSON"
	invalidJSON2           = "{\"a\": 5, \"b\": \"c\": 6}}"
	mixedValidityJSON      = "[{\"a\": 5, \"time\": \"2006-01-02T15:04:05\"}, {\"a\": 2}]"
)

const validJSONTags = `
{
    "a": 5,
    "b": {
      "c": 6
    },
    "mytag": "foobar",
    "othertag": "baz",
    "tags_object": {
        "mytag": "foobar",
        "othertag": "baz"
    }
}
`

const validJSONArrayTags = `
[
{
    "a": 5,
    "b": {
        "c": 6
    },
    "mytag": "foo",
    "othertag": "baz",
    "tags_array": [
        {
        "mytag": "foo"
        },
        {
        "othertag": "baz"
        }
    ],
    "anothert": "foo"
},
{
    "a": 7,
    "b": {
        "c": 8
    },
    "mytag": "bar",
    "othertag": "baz",
    "tags_array": [
    {
    "mytag": "bar"
    },
    {
    "othertag": "baz"
    }
    ],
    "anothert": "bar"
    }
]
`

const benchmarkData = `
[
	{
	"name": "impression",
	"timestamp": 1653643420,
	"fields": {
		"count_sum": 5
	},
	"tags": {
		"key": "12345",
		"flagname": "F5",
		"host": "1cbbb3796fc2",
		"platform": "Java",
		"sdkver": "4.9.1",
		"value": "false"
	}
	},
	{
	"name": "expression",
	"timestamp": 1653646789,
	"fields": {
		"count_sum": 42
	},
	"tags": {
		"key": "67890",
		"flagname": "E42",
		"host": "klaus",
		"platform": "Golang",
		"sdkver": "1.18.3",
		"value": "true"
	}
	}
]`

func TestParseValidJSON(t *testing.T) {
	parser := &Parser{MetricName: "json_test"}
	require.NoError(t, parser.Init())

	// Most basic vanilla test
	actual, err := parser.Parse([]byte(validJSON))
	require.NoError(t, err)
	require.Len(t, actual, 1)
	require.Equal(t, "json_test", actual[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, actual[0].Fields())
	require.Equal(t, map[string]string{}, actual[0].Tags())

	// Test that newlines are fine
	actual, err = parser.Parse([]byte(validJSONNewline))
	require.NoError(t, err)
	require.Len(t, actual, 1)
	require.Equal(t, "json_test", actual[0].Name())
	require.Equal(t, map[string]interface{}{
		"d":   float64(7),
		"b_d": float64(8),
	}, actual[0].Fields())
	require.Equal(t, map[string]string{}, actual[0].Tags())

	// Test that strings without TagKeys defined are ignored
	actual, err = parser.Parse([]byte(validJSONTags))
	require.NoError(t, err)
	require.Len(t, actual, 1)
	require.Equal(t, "json_test", actual[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, actual[0].Fields())
	require.Equal(t, map[string]string{}, actual[0].Tags())

	// Test that whitespace only will parse as an empty list of actual
	actual, err = parser.Parse([]byte("\n\t"))
	require.NoError(t, err)
	require.Empty(t, actual)

	// Test that an empty string will parse as an empty list of actual
	actual, err = parser.Parse([]byte(""))
	require.NoError(t, err)
	require.Empty(t, actual)
}

func TestParseLineValidJSON(t *testing.T) {
	parser := &Parser{MetricName: "json_test"}
	require.NoError(t, parser.Init())

	// Most basic vanilla test
	actual, err := parser.ParseLine(validJSON)
	require.NoError(t, err)
	require.Equal(t, "json_test", actual.Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, actual.Fields())
	require.Equal(t, map[string]string{}, actual.Tags())

	// Test that newlines are fine
	actual, err = parser.ParseLine(validJSONNewline)
	require.NoError(t, err)
	require.Equal(t, "json_test", actual.Name())
	require.Equal(t, map[string]interface{}{
		"d":   float64(7),
		"b_d": float64(8),
	}, actual.Fields())
	require.Equal(t, map[string]string{}, actual.Tags())

	// Test that strings without TagKeys defined are ignored
	actual, err = parser.ParseLine(validJSONTags)
	require.NoError(t, err)
	require.Equal(t, "json_test", actual.Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, actual.Fields())
	require.Equal(t, map[string]string{}, actual.Tags())
}

func TestParseInvalidJSON(t *testing.T) {
	parser := &Parser{MetricName: "json_test"}
	require.NoError(t, parser.Init())

	_, err := parser.Parse([]byte(invalidJSON))
	require.Error(t, err)
	_, err = parser.Parse([]byte(invalidJSON2))
	require.Error(t, err)
	_, err = parser.ParseLine(invalidJSON)
	require.Error(t, err)
}

func TestParseJSONImplicitStrictness(t *testing.T) {
	parserImplicitNoStrict := &Parser{
		MetricName: "json_test",
		TimeKey:    "time",
	}
	require.NoError(t, parserImplicitNoStrict.Init())

	_, err := parserImplicitNoStrict.Parse([]byte(mixedValidityJSON))
	require.NoError(t, err)
}

func TestParseJSONExplicitStrictnessFalse(t *testing.T) {
	parserNoStrict := &Parser{
		MetricName: "json_test",
		TimeKey:    "time",
		Strict:     false,
	}
	require.NoError(t, parserNoStrict.Init())

	_, err := parserNoStrict.Parse([]byte(mixedValidityJSON))
	require.NoError(t, err)
}

func TestParseJSONExplicitStrictnessTrue(t *testing.T) {
	parserStrict := &Parser{
		MetricName: "json_test",
		TimeKey:    "time",
		Strict:     true,
	}
	require.NoError(t, parserStrict.Init())

	_, err := parserStrict.Parse([]byte(mixedValidityJSON))
	require.Error(t, err)
}

func TestParseWithTagKeys(t *testing.T) {
	// Test that strings not matching tag keys are ignored
	parser := &Parser{
		MetricName: "json_test",
		TagKeys:    []string{"wrongtagkey"},
	}
	require.NoError(t, parser.Init())

	actual, err := parser.Parse([]byte(validJSONTags))
	require.NoError(t, err)
	require.Len(t, actual, 1)
	require.Equal(t, "json_test", actual[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, actual[0].Fields())
	require.Equal(t, map[string]string{}, actual[0].Tags())

	// Test that single tag key is found and applied
	parser = &Parser{
		MetricName: "json_test",
		TagKeys:    []string{"mytag"},
	}
	require.NoError(t, parser.Init())

	actual, err = parser.Parse([]byte(validJSONTags))
	require.NoError(t, err)
	require.Len(t, actual, 1)
	require.Equal(t, "json_test", actual[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, actual[0].Fields())
	require.Equal(t, map[string]string{
		"mytag": "foobar",
	}, actual[0].Tags())

	// Test that both tag keys are found and applied
	parser = &Parser{
		MetricName: "json_test",
		TagKeys:    []string{"mytag", "othertag"},
	}
	require.NoError(t, parser.Init())

	actual, err = parser.Parse([]byte(validJSONTags))
	require.NoError(t, err)
	require.Len(t, actual, 1)
	require.Equal(t, "json_test", actual[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, actual[0].Fields())
	require.Equal(t, map[string]string{
		"mytag":    "foobar",
		"othertag": "baz",
	}, actual[0].Tags())
}

func TestParseLineWithTagKeys(t *testing.T) {
	// Test that strings not matching tag keys are ignored
	parser := &Parser{
		MetricName: "json_test",
		TagKeys:    []string{"wrongtagkey"},
	}
	require.NoError(t, parser.Init())

	actual, err := parser.ParseLine(validJSONTags)
	require.NoError(t, err)
	require.Equal(t, "json_test", actual.Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, actual.Fields())
	require.Equal(t, map[string]string{}, actual.Tags())

	// Test that single tag key is found and applied
	parser = &Parser{
		MetricName: "json_test",
		TagKeys:    []string{"mytag"},
	}
	require.NoError(t, parser.Init())

	actual, err = parser.ParseLine(validJSONTags)
	require.NoError(t, err)
	require.Equal(t, "json_test", actual.Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, actual.Fields())
	require.Equal(t, map[string]string{
		"mytag": "foobar",
	}, actual.Tags())

	// Test that both tag keys are found and applied
	parser = &Parser{
		MetricName: "json_test",
		TagKeys:    []string{"mytag", "othertag"},
	}
	require.NoError(t, parser.Init())

	actual, err = parser.ParseLine(validJSONTags)
	require.NoError(t, err)
	require.Equal(t, "json_test", actual.Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, actual.Fields())
	require.Equal(t, map[string]string{
		"mytag":    "foobar",
		"othertag": "baz",
	}, actual.Tags())
}

func TestParseValidJSONDefaultTags(t *testing.T) {
	parser := &Parser{
		MetricName:  "json_test",
		TagKeys:     []string{"mytag"},
		DefaultTags: map[string]string{"t4g": "default"},
	}
	require.NoError(t, parser.Init())

	// Most basic vanilla test
	actual, err := parser.Parse([]byte(validJSON))
	require.NoError(t, err)
	require.Len(t, actual, 1)
	require.Equal(t, "json_test", actual[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, actual[0].Fields())
	require.Equal(t, map[string]string{"t4g": "default"}, actual[0].Tags())

	// Test that tagkeys and default tags are applied
	actual, err = parser.Parse([]byte(validJSONTags))
	require.NoError(t, err)
	require.Len(t, actual, 1)
	require.Equal(t, "json_test", actual[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, actual[0].Fields())
	require.Equal(t, map[string]string{
		"t4g":   "default",
		"mytag": "foobar",
	}, actual[0].Tags())
}

// Test that default tags are overridden by tag keys
func TestParseValidJSONDefaultTagsOverride(t *testing.T) {
	parser := &Parser{
		MetricName:  "json_test",
		TagKeys:     []string{"mytag"},
		DefaultTags: map[string]string{"mytag": "default"},
	}
	require.NoError(t, parser.Init())

	// Most basic vanilla test
	actual, err := parser.Parse([]byte(validJSON))
	require.NoError(t, err)
	require.Len(t, actual, 1)
	require.Equal(t, "json_test", actual[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, actual[0].Fields())
	require.Equal(t, map[string]string{"mytag": "default"}, actual[0].Tags())

	// Test that tagkeys override default tags
	actual, err = parser.Parse([]byte(validJSONTags))
	require.NoError(t, err)
	require.Len(t, actual, 1)
	require.Equal(t, "json_test", actual[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, actual[0].Fields())
	require.Equal(t, map[string]string{
		"mytag": "foobar",
	}, actual[0].Tags())
}

// Test that json arrays can be parsed
func TestParseValidJSONArray(t *testing.T) {
	parser := &Parser{MetricName: "json_array_test"}
	require.NoError(t, parser.Init())

	// Most basic vanilla test
	actual, err := parser.Parse([]byte(validJSONArray))
	require.NoError(t, err)
	require.Len(t, actual, 1)
	require.Equal(t, "json_array_test", actual[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, actual[0].Fields())
	require.Equal(t, map[string]string{}, actual[0].Tags())

	// Basic multiple datapoints
	actual, err = parser.Parse([]byte(validJSONArrayMultiple))
	require.NoError(t, err)
	require.Len(t, actual, 2)
	require.Equal(t, "json_array_test", actual[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, actual[0].Fields())
	require.Equal(t, map[string]string{}, actual[1].Tags())
	require.Equal(t, "json_array_test", actual[1].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(7),
		"b_c": float64(8),
	}, actual[1].Fields())
	require.Equal(t, map[string]string{}, actual[1].Tags())
}

func TestParseArrayWithTagKeys(t *testing.T) {
	// Test that strings not matching tag keys are ignored
	parser := &Parser{
		MetricName: "json_array_test",
		TagKeys:    []string{"wrongtagkey"},
	}
	require.NoError(t, parser.Init())

	actual, err := parser.Parse([]byte(validJSONArrayTags))
	require.NoError(t, err)
	require.Len(t, actual, 2)
	require.Equal(t, "json_array_test", actual[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, actual[0].Fields())
	require.Equal(t, map[string]string{}, actual[0].Tags())

	require.Equal(t, "json_array_test", actual[1].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(7),
		"b_c": float64(8),
	}, actual[1].Fields())
	require.Equal(t, map[string]string{}, actual[1].Tags())

	// Test that single tag key is found and applied
	parser = &Parser{
		MetricName: "json_array_test",
		TagKeys:    []string{"mytag"},
	}
	require.NoError(t, parser.Init())

	actual, err = parser.Parse([]byte(validJSONArrayTags))
	require.NoError(t, err)
	require.Len(t, actual, 2)
	require.Equal(t, "json_array_test", actual[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, actual[0].Fields())
	require.Equal(t, map[string]string{
		"mytag": "foo",
	}, actual[0].Tags())

	require.Equal(t, "json_array_test", actual[1].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(7),
		"b_c": float64(8),
	}, actual[1].Fields())
	require.Equal(t, map[string]string{
		"mytag": "bar",
	}, actual[1].Tags())

	// Test that both tag keys are found and applied
	parser = &Parser{
		MetricName: "json_array_test",
		TagKeys:    []string{"mytag", "othertag"},
	}
	require.NoError(t, parser.Init())

	actual, err = parser.Parse([]byte(validJSONArrayTags))
	require.NoError(t, err)
	require.Len(t, actual, 2)
	require.Equal(t, "json_array_test", actual[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, actual[0].Fields())
	require.Equal(t, map[string]string{
		"mytag":    "foo",
		"othertag": "baz",
	}, actual[0].Tags())

	require.Equal(t, "json_array_test", actual[1].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(7),
		"b_c": float64(8),
	}, actual[1].Fields())
	require.Equal(t, map[string]string{
		"mytag":    "bar",
		"othertag": "baz",
	}, actual[1].Tags())
}

var jsonBOM = []byte("\xef\xbb\xbf[{\"value\":17}]")

func TestHttpJsonBOM(t *testing.T) {
	parser := &Parser{MetricName: "json_test"}
	require.NoError(t, parser.Init())

	// Most basic vanilla test
	_, err := parser.Parse(jsonBOM)
	require.NoError(t, err)
}

// for testing issue #4260
func TestJSONParseNestedArray(t *testing.T) {
	testString := `{
	"total_devices": 5,
	"total_threads": 10,
	"shares": {
		"total": 5,
		"accepted": 5,
		"rejected": 0,
		"avg_find_time": 4,
		"tester": "work",
		"tester2": "don't want this",
		"tester3": {
			"hello":"sup",
			"fun":"money",
			"break":9
		}
	}
	}`

	parser := &Parser{
		MetricName: "json_test",
		TagKeys:    []string{"total_devices", "total_threads", "shares_tester3_fun"},
	}
	require.NoError(t, parser.Init())

	actual, err := parser.Parse([]byte(testString))
	require.Len(t, actual, 1)
	require.NoError(t, err)
	require.Len(t, actual[0].Tags(), 3)
}

func TestJSONQueryErrorOnArray(t *testing.T) {
	testString := `{
		"total_devices": 5,
		"total_threads": 10,
		"shares": {
			"total": 5,
			"accepted": 6,
			"test_string": "don't want this",
			"test_obj": {
				"hello":"sup",
				"fun":"money",
				"break":9
			},
			"myArr":[4,5,6]
		}
	}`

	parser := &Parser{
		MetricName: "json_test",
		Query:      "shares.myArr",
	}
	require.NoError(t, parser.Init())

	_, err := parser.Parse([]byte(testString))
	require.Error(t, err)
}

func TestArrayOfObjects(t *testing.T) {
	testString := `{
		"meta": {
			"info":9,
			"shares": [{
				"channel": 6,
				"time": 1130,
				"ice":"man"
			},
			{
				"channel": 5,
				"time": 1030,
				"ice":"bucket"
			},
			{
				"channel": 10,
				"time": 330,
				"ice":"cream"
			}]
		},
		"more_stuff":"junk"
	}`

	parser := &Parser{
		MetricName: "json_test",
		TagKeys:    []string{"ice"},
		Query:      "meta.shares",
	}
	require.NoError(t, parser.Init())

	actual, err := parser.Parse([]byte(testString))
	require.NoError(t, err)
	require.Len(t, actual, 3)
}

func TestUseCaseJSONQuery(t *testing.T) {
	testString := `{
		"obj": {
			"name": {"first": "Tom", "last": "Anderson"},
			"age":37,
			"children": ["Sara","Alex","Jack"],
			"fav.movie": "Deer Hunter",
			"friends": [
				{"first": "Dale", "last": "Murphy", "age": 44},
				{"first": "Roger", "last": "Craig", "age": 68},
				{"first": "Jane", "last": "Murphy", "age": 47}
			]
		}
	}`

	parser := &Parser{
		MetricName:   "json_test",
		StringFields: []string{"last"},
		TagKeys:      []string{"first"},
		Query:        "obj.friends",
	}
	require.NoError(t, parser.Init())

	actual, err := parser.Parse([]byte(testString))
	require.NoError(t, err)
	require.Len(t, actual, 3)
	require.Equal(t, "Murphy", actual[0].Fields()["last"])
}

func TestTimeParser(t *testing.T) {
	testString := `[
		{
			"a": 5,
			"b": {
				"c": 6,
				"time":"04 Jan 06 15:04 MST"
			},
			"my_tag_1": "foo",
			"my_tag_2": "baz"
		},
		{
			"a": 7,
			"b": {
				"c": 8,
				"time":"11 Jan 07 15:04 MST"
			},
			"my_tag_1": "bar",
			"my_tag_2": "baz"
		}
	]`

	parser := &Parser{
		MetricName: "json_test",
		TimeKey:    "b_time",
		TimeFormat: "02 Jan 06 15:04 MST",
	}
	require.NoError(t, parser.Init())

	actual, err := parser.Parse([]byte(testString))
	require.NoError(t, err)
	require.Len(t, actual, 2)
	require.NotEqual(t, actual[0].Time(), actual[1].Time())
}

func TestTimeParserWithTimezone(t *testing.T) {
	testString := `{
		"time": "04 Jan 06 15:04"
	}`

	parser := &Parser{
		MetricName: "json_test",
		TimeKey:    "time",
		TimeFormat: "02 Jan 06 15:04",
		Timezone:   "America/New_York",
	}
	require.NoError(t, parser.Init())

	actual, err := parser.Parse([]byte(testString))
	require.NoError(t, err)
	require.Len(t, actual, 1)
	require.EqualValues(t, int64(1136405040000000000), actual[0].Time().UnixNano())
}

func TestUnixTimeParser(t *testing.T) {
	testString := `[
		{
			"a": 5,
			"b": {
				"c": 6,
				"time": "1536001411.1234567890"
			},
			"my_tag_1": "foo",
			"my_tag_2": "baz"
		},
		{
			"a": 7,
			"b": {
				"c": 8,
				"time": 1536002769.123
			},
			"my_tag_1": "bar",
			"my_tag_2": "baz"
		}
	]`

	parser := &Parser{
		MetricName: "json_test",
		TimeKey:    "b_time",
		TimeFormat: "unix",
	}
	require.NoError(t, parser.Init())

	actual, err := parser.Parse([]byte(testString))
	require.NoError(t, err)
	require.Len(t, actual, 2)
	require.NotEqual(t, actual[0].Time(), actual[1].Time())
}

func TestUnixMsTimeParser(t *testing.T) {
	testString := `[
		{
			"a": 5,
			"b": {
				"c": 6,
				"time": "1536001411100"
			},
			"my_tag_1": "foo",
			"my_tag_2": "baz"
		},
		{
			"a": 7,
			"b": {
				"c": 8,
				"time": 1536002769123
			},
			"my_tag_1": "bar",
			"my_tag_2": "baz"
		}
	]`

	parser := &Parser{
		MetricName: "json_test",
		TimeKey:    "b_time",
		TimeFormat: "unix_ms",
	}
	require.NoError(t, parser.Init())

	actual, err := parser.Parse([]byte(testString))
	require.NoError(t, err)
	require.Len(t, actual, 2)
	require.NotEqual(t, actual[0].Time(), actual[1].Time())
}

func TestTimeErrors(t *testing.T) {
	testString := `{
		"a": 5,
		"b": {
			"c": 6,
			"time":"04 Jan 06 15:04 MST"
		},
		"my_tag_1": "foo",
		"my_tag_2": "baz"
	}`

	parser := &Parser{
		MetricName: "json_test",
		TimeKey:    "b_time",
		TimeFormat: "02 January 06 15:04 MST",
	}
	require.NoError(t, parser.Init())

	actual, err := parser.Parse([]byte(testString))
	require.Error(t, err)
	require.Empty(t, actual)

	testString2 := `{
		"a": 5,
		"b": {
			"c": 6
		},
		"my_tag_1": "foo",
		"my_tag_2": "baz"
	}`

	parser = &Parser{
		MetricName: "json_test",
		TimeKey:    "b_time",
		TimeFormat: "02 January 06 15:04 MST",
	}
	require.NoError(t, parser.Init())

	actual, err = parser.Parse([]byte(testString2))
	require.Error(t, err)
	require.Empty(t, actual)
	require.Equal(t, errors.New("'json_time_key' could not be found"), err)
}

func TestShareTimestamp(t *testing.T) {
	parser := &Parser{MetricName: "json_test"}
	require.NoError(t, parser.Init())

	actual, err := parser.Parse([]byte(validJSONArrayMultiple))
	require.NoError(t, err)
	require.Len(t, actual, 2)
	require.Equal(t, actual[0].Time(), actual[1].Time())
}

func TestNameKey(t *testing.T) {
	testString := `{
		"a": 5,
		"b": {
			"c": "this is my name",
			"time":"04 Jan 06 15:04 MST"
		},
		"my_tag_1": "foo",
		"my_tag_2": "baz"
	}`

	parser := &Parser{NameKey: "b_c"}
	require.NoError(t, parser.Init())

	actual, err := parser.Parse([]byte(testString))
	require.NoError(t, err)
	require.Equal(t, "this is my name", actual[0].Name())
}

func TestParseArrayWithWrongType(t *testing.T) {
	data := `[{"answer": 42}, 123]`

	parser := &Parser{}
	require.NoError(t, parser.Init())

	_, err := parser.Parse([]byte(data))
	require.Error(t, err)
}

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		parser   *Parser
		input    []byte
		expected []Dana.Metric
	}{
		{
			name: "tag keys with underscore issue 6705",
			parser: &Parser{
				MetricName: "json",
				TagKeys:    []string{"metric___name__"},
			},
			input: []byte(`{"metric": {"__name__": "howdy", "time_idle": 42}}`),
			expected: []Dana.Metric{
				testutil.MustMetric(
					"json",
					map[string]string{
						"metric___name__": "howdy",
					},
					map[string]interface{}{
						"metric_time_idle": 42.0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name:   "parse empty array",
			parser: &Parser{},
			input:  []byte(`[]`),
		},
		{
			name:   "parse null",
			parser: &Parser{},
			input:  []byte(`null`),
		},
		{
			name:   "parse null with query",
			parser: &Parser{Query: "result.data"},
			input:  []byte(`{"error":null,"result":{"data":null,"items_per_page":10,"total_items":0,"total_pages":0}}`),
		},
		{
			name: "parse simple array",
			parser: &Parser{
				MetricName: "json",
			},
			input: []byte(`[{"answer": 42}]`),
			expected: []Dana.Metric{
				testutil.MustMetric(
					"json",
					map[string]string{},
					map[string]interface{}{
						"answer": 42.0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "string field glob",
			parser: &Parser{
				MetricName:   "json",
				StringFields: []string{"*"},
			},
			input: []byte(`
{
    "color": "red",
    "status": "error"
}
`),
			expected: []Dana.Metric{
				testutil.MustMetric(
					"json",
					map[string]string{},
					map[string]interface{}{
						"color":  "red",
						"status": "error",
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "time key is deleted from fields",
			parser: &Parser{
				MetricName: "json",
				TimeKey:    "timestamp",
				TimeFormat: "unix",
			},
			input: []byte(`
{
	"value": 42,
	"timestamp":  1541183052
}
`),
			expected: []Dana.Metric{
				testutil.MustMetric(
					"json",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(1541183052, 0),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := tt.parser
			require.NoError(t, parser.Init())

			actual, err := parser.Parse(tt.input)
			require.NoError(t, err)

			testutil.RequireMetricsEqual(t, tt.expected, actual, testutil.IgnoreTime())
		})
	}
}

func TestParseWithWildcardTagKeys(t *testing.T) {
	var tests = []struct {
		name     string
		parser   *Parser
		input    []byte
		expected []Dana.Metric
	}{
		{
			name: "wildcard matching with tags nested within object",
			parser: &Parser{
				MetricName: "json_test",
				TagKeys:    []string{"tags_object_*"},
			},
			input: []byte(validJSONTags),
			expected: []Dana.Metric{
				testutil.MustMetric(
					"json_test",
					map[string]string{
						"tags_object_mytag":    "foobar",
						"tags_object_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(5),
						"b_c": float64(6),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "wildcard matching with keys containing tag",
			parser: &Parser{
				MetricName: "json_test",
				TagKeys:    []string{"*tag"},
			},
			input: []byte(validJSONTags),
			expected: []Dana.Metric{
				testutil.MustMetric(
					"json_test",
					map[string]string{
						"mytag":                "foobar",
						"othertag":             "baz",
						"tags_object_mytag":    "foobar",
						"tags_object_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(5),
						"b_c": float64(6),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "strings not matching tag keys are still also ignored",
			parser: &Parser{
				MetricName: "json_test",
				TagKeys:    []string{"wrongtagkey", "tags_object_*"},
			},
			input: []byte(validJSONTags),
			expected: []Dana.Metric{
				testutil.MustMetric(
					"json_test",
					map[string]string{
						"tags_object_mytag":    "foobar",
						"tags_object_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(5),
						"b_c": float64(6),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "single tag key is also found and applied",
			parser: &Parser{
				MetricName: "json_test",
				TagKeys:    []string{"mytag", "tags_object_*"},
			},
			input: []byte(validJSONTags),
			expected: []Dana.Metric{
				testutil.MustMetric(
					"json_test",
					map[string]string{
						"mytag":                "foobar",
						"tags_object_mytag":    "foobar",
						"tags_object_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(5),
						"b_c": float64(6),
					},
					time.Unix(0, 0),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := tt.parser
			require.NoError(t, parser.Init())

			actual, err := parser.Parse(tt.input)
			require.NoError(t, err)
			testutil.RequireMetricsEqual(t, tt.expected, actual, testutil.IgnoreTime())
		})
	}
}

func TestParseLineWithWildcardTagKeys(t *testing.T) {
	var tests = []struct {
		name     string
		parser   *Parser
		input    string
		expected Dana.Metric
	}{
		{
			name: "wildcard matching with tags nested within object",
			parser: &Parser{
				MetricName: "json_test",
				TagKeys:    []string{"tags_object_*"},
			},
			input: validJSONTags,
			expected: testutil.MustMetric(
				"json_test",
				map[string]string{
					"tags_object_mytag":    "foobar",
					"tags_object_othertag": "baz",
				},
				map[string]interface{}{
					"a":   float64(5),
					"b_c": float64(6),
				},
				time.Unix(0, 0),
			),
		},
		{
			name: "wildcard matching with keys containing tag",
			parser: &Parser{
				MetricName: "json_test",
				TagKeys:    []string{"*tag"},
			},
			input: validJSONTags,
			expected: testutil.MustMetric(
				"json_test",
				map[string]string{
					"mytag":                "foobar",
					"othertag":             "baz",
					"tags_object_mytag":    "foobar",
					"tags_object_othertag": "baz",
				},
				map[string]interface{}{
					"a":   float64(5),
					"b_c": float64(6),
				},
				time.Unix(0, 0),
			),
		},
		{
			name: "strings not matching tag keys are ignored",
			parser: &Parser{
				MetricName: "json_test",
				TagKeys:    []string{"wrongtagkey", "tags_object_*"},
			},
			input: validJSONTags,
			expected: testutil.MustMetric(
				"json_test",
				map[string]string{
					"tags_object_mytag":    "foobar",
					"tags_object_othertag": "baz",
				},
				map[string]interface{}{
					"a":   float64(5),
					"b_c": float64(6),
				},
				time.Unix(0, 0),
			),
		},
		{
			name: "single tag key is also found and applied",
			parser: &Parser{
				MetricName: "json_test",
				TagKeys:    []string{"mytag", "tags_object_*"},
			},
			input: validJSONTags,
			expected: testutil.MustMetric(
				"json_test",
				map[string]string{
					"mytag":                "foobar",
					"tags_object_mytag":    "foobar",
					"tags_object_othertag": "baz",
				},
				map[string]interface{}{
					"a":   float64(5),
					"b_c": float64(6),
				},
				time.Unix(0, 0),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := tt.parser
			require.NoError(t, parser.Init())

			actual, err := parser.ParseLine(tt.input)
			require.NoError(t, err)

			testutil.RequireMetricEqual(t, tt.expected, actual, testutil.IgnoreTime())
		})
	}
}

func TestParseArrayWithWildcardTagKeys(t *testing.T) {
	var tests = []struct {
		name     string
		parser   *Parser
		input    []byte
		expected []Dana.Metric
	}{
		{
			name: "wildcard matching with keys containing tag within array works",
			parser: &Parser{
				MetricName: "json_array_test",
				TagKeys:    []string{"*tag"},
			},
			input: []byte(validJSONArrayTags),
			expected: []Dana.Metric{
				testutil.MustMetric(
					"json_array_test",
					map[string]string{
						"mytag":                 "foo",
						"othertag":              "baz",
						"tags_array_0_mytag":    "foo",
						"tags_array_1_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(5),
						"b_c": float64(6),
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"json_array_test",
					map[string]string{
						"mytag":                 "bar",
						"othertag":              "baz",
						"tags_array_0_mytag":    "bar",
						"tags_array_1_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(7),
						"b_c": float64(8),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: " wildcard matching with tags nested array within object works",
			parser: &Parser{
				MetricName: "json_array_test",
				TagKeys:    []string{"tags_array_*"},
			},
			input: []byte(validJSONArrayTags),
			expected: []Dana.Metric{
				testutil.MustMetric(
					"json_array_test",
					map[string]string{
						"tags_array_0_mytag":    "foo",
						"tags_array_1_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(5),
						"b_c": float64(6),
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"json_array_test",
					map[string]string{
						"tags_array_0_mytag":    "bar",
						"tags_array_1_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(7),
						"b_c": float64(8),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "strings not matching tag keys are still also ignored",
			parser: &Parser{
				MetricName: "json_array_test",
				TagKeys:    []string{"mytag", "*tag"},
			},
			input: []byte(validJSONArrayTags),
			expected: []Dana.Metric{
				testutil.MustMetric(
					"json_array_test",
					map[string]string{
						"mytag":                 "foo",
						"othertag":              "baz",
						"tags_array_0_mytag":    "foo",
						"tags_array_1_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(5),
						"b_c": float64(6),
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"json_array_test",
					map[string]string{
						"mytag":                 "bar",
						"othertag":              "baz",
						"tags_array_0_mytag":    "bar",
						"tags_array_1_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(7),
						"b_c": float64(8),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "single tag key is also found and applied",
			parser: &Parser{
				MetricName: "json_array_test",
				TagKeys:    []string{"anothert", "*tag"},
			},
			input: []byte(validJSONArrayTags),
			expected: []Dana.Metric{
				testutil.MustMetric(
					"json_array_test",
					map[string]string{
						"anothert":              "foo",
						"mytag":                 "foo",
						"othertag":              "baz",
						"tags_array_0_mytag":    "foo",
						"tags_array_1_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(5),
						"b_c": float64(6),
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"json_array_test",
					map[string]string{
						"anothert":              "bar",
						"mytag":                 "bar",
						"othertag":              "baz",
						"tags_array_0_mytag":    "bar",
						"tags_array_1_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(7),
						"b_c": float64(8),
					},
					time.Unix(0, 0),
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := tt.parser
			require.NoError(t, parser.Init())

			actual, err := parser.Parse(tt.input)
			require.NoError(t, err)

			testutil.RequireMetricsEqual(t, tt.expected, actual, testutil.IgnoreTime())
		})
	}
}

func TestBenchmarkData(t *testing.T) {
	// Setup the plugin
	plugin := &Parser{
		MetricName: "benchmark",
		TagKeys:    []string{"tags_*"},
	}
	require.NoError(t, plugin.Init())

	expected := []Dana.Metric{
		metric.New(
			"benchmark",
			map[string]string{
				"tags_flagname": "F5",
				"tags_host":     "1cbbb3796fc2",
				"tags_key":      "12345",
				"tags_platform": "Java",
				"tags_sdkver":   "4.9.1",
				"tags_value":    "false",
			},
			map[string]interface{}{
				"fields_count_sum": float64(5),
				"timestamp":        float64(1653643420),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"benchmark",
			map[string]string{
				"tags_flagname": "E42",
				"tags_host":     "klaus",
				"tags_key":      "67890",
				"tags_platform": "Golang",
				"tags_sdkver":   "1.18.3",
				"tags_value":    "true",
			},
			map[string]interface{}{
				"fields_count_sum": float64(42),
				"timestamp":        float64(1653646789),
			},
			time.Unix(0, 0),
		),
	}

	// Do the parsing
	actual, err := plugin.Parse([]byte(benchmarkData))
	require.NoError(t, err)
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
}

func BenchmarkParsingSequential(b *testing.B) {
	// Configure the plugin
	plugin := &Parser{
		MetricName: "benchmark",
		TagKeys:    []string{"tags_*"},
	}
	require.NoError(b, plugin.Init())

	// Do the benchmarking
	for n := 0; n < b.N; n++ {
		//nolint:errcheck // Benchmarking so skip the error check to avoid the unnecessary operations
		plugin.Parse([]byte(benchmarkData))
	}
}

func BenchmarkParsingParallel(b *testing.B) {
	// Configure the plugin
	plugin := &Parser{
		MetricName: "benchmark",
		TagKeys:    []string{"tags_*"},
	}
	require.NoError(b, plugin.Init())

	// Do the benchmarking
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			//nolint:errcheck // Benchmarking so skip the error check to avoid the unnecessary operations
			plugin.Parse([]byte(benchmarkData))
		}
	})
}

func FuzzParserJSON(f *testing.F) {
	for _, value := range fuzz.JSONDictionary {
		f.Add([]byte(value))
	}

	f.Add([]byte(validJSON))
	f.Add([]byte(validJSONArray))
	f.Add([]byte(validJSONArrayMultiple))
	f.Add([]byte(validJSONArrayTags))
	f.Add([]byte(validJSONNewline))
	f.Add([]byte(validJSONTags))

	parser := &Parser{MetricName: "testing"}
	require.NoError(f, parser.Init())

	f.Fuzz(func(_ *testing.T, input []byte) {
		//nolint:errcheck // fuzz testing can give lots of errors, but we just want to test for crashes
		parser.Parse(input)
	})
}
