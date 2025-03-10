package media

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"
)

var (
	defaultValidFields      = []string{"name", "extension", "nodeId"}
	canSortFieldByExtension = map[string][]string{
		"mp3": {"album", "title", "artist"},
	}
)

/* Takes both items being compared and returns a function that can be used to sort on another property.
** If no additional sorting is valid, returns no function.
** This function assumes both values of the sortBy are equal hence we need additional sorting.
 */
func getEqualValueSorter(item1, item2 map[string]any, sortBy string) func() bool {
	extension1, ok := item1["extension"].(string)
	if !ok {
		panic(errors.New("invalid item, does not contain a proper extension"))
	}
	extension2, ok := item2["extension"].(string)
	if !ok {
		panic(errors.New("invalid item, does not contain a proper extension"))
	}

	// ensure both items are the same extension
	if extension1 != extension2 {
		return nil
	}

	// if item is an mp3 and we are sorting by album, return name in alphabetical order
	if extension1 == "mp3" && sortBy == "album" {
		return func() bool {
			name1 := item1["name"].(string)
			name2 := item2["name"].(string)

			return name1 < name2
		}
	}

	return nil
}

func canSort(fieldName, extension string, validFields []string) error {
	valid := false
	for _, validField := range validFields {
		if validField == fieldName {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf(
			"invalid sortBy field %s, items with extension %s can only be sorted by %s",
			fieldName,
			extension,
			strings.Join(validFields, ", "),
		)
	}

	return nil
}

func canSortField(item map[string]any, sortBy string) error {
	extension, ok := item["extension"].(string)
	if !ok {
		return errors.New("invalid item, does not contain a proper extension")
	}

	validSortFields := make([]string, 0)
	validSortFields = append(validSortFields, defaultValidFields...)
	fields, ok := canSortFieldByExtension[extension]
	if ok {
		validSortFields = append(validSortFields, fields...)
	}

	return canSort(sortBy, extension, validSortFields)
}

func sortMedia(media []map[string]any, sortBy, order string) (err error) {
	// check if all items can be sorted by this field
	for _, m := range media {
		sErr := canSortField(m, sortBy)
		if sErr != nil {
			return sErr
		}
	}

	// sort.Slice does not allow for error handling, therefore panics will be used and recovered to be returned as an error
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				log.Error().Msgf("internal panic in the sort %v", r)
				err = errors.New("an error occured when sorting media")
			}
		}
	}()

	sort.Slice(media, func(i, j int) bool {
		val1 := reflect.ValueOf(media[i][sortBy])
		val2 := reflect.ValueOf(media[j][sortBy])

		// if empty, look for field in metadata
		if !val1.IsValid() {
			metadata1, ok := media[i]["metadata"].(map[string]interface{})
			if !ok {
				panic(errors.New("invalid media item, metadata has invalid format"))
			}

			val1 = reflect.ValueOf(metadata1[sortBy])
			if val1.IsZero() {
				return false
			}

			metadata2, ok := media[j]["metadata"].(map[string]interface{})
			if !ok {
				panic(errors.New("invalid media item, metadata has invalid format"))
			}

			val2 = reflect.ValueOf(metadata2[sortBy])
			if val2.IsZero() {
				return false
			}
		}

		if order == "asc" {
			switch val1.Kind() {
			case reflect.String:
				if val1.String() == val2.String() {
					eqValSorter := getEqualValueSorter(media[i], media[j], sortBy)
					if eqValSorter != nil {
						return eqValSorter()
					}
				}

				return val1.String() < val2.String()
			case reflect.Int:
				if val1.Int() == val2.Int() {
					eqValSorter := getEqualValueSorter(media[i], media[j], sortBy)
					if eqValSorter != nil {
						return eqValSorter()
					}
				}

				return val1.Int() < val2.Int()
			default:
				panic(fmt.Errorf("field %s is not an expected format", sortBy))
			}
		} else {
			switch val1.Kind() {
			case reflect.String:
				if val1.String() == val2.String() {
					eqValSorter := getEqualValueSorter(media[i], media[j], sortBy)
					if eqValSorter != nil {
						return eqValSorter()
					}
				}

				return val1.String() > val2.String()
			case reflect.Int:
				if val1.Int() == val2.Int() {
					eqValSorter := getEqualValueSorter(media[i], media[j], sortBy)
					if eqValSorter != nil {
						return eqValSorter()
					}
				}

				return val1.Int() > val2.Int()
			default:
				panic(fmt.Errorf("field %s is not an expected format", sortBy))
			}
		}
	})

	return
}
