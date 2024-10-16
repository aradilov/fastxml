package fastxml

import (
	"bytes"
	"fmt"
	cedar "github.com/aradilov/ahocorasick"
	"github.com/orisano/gosax"
)

type Iterator func(name, value, path []byte, attributes Attributes) error

type XMLReader struct {
	r        *gosax.Reader
	register *cedar.Cedar
}

func NewReader(bb []byte) *XMLReader {
	reader := bytes.NewReader(bb)
	r := gosax.NewReader(reader)

	return &XMLReader{r: r}
}

func (xml *XMLReader) Path(path string, cb Iterator) error {
	if len(path) == 0 {
		return fmt.Errorf("path is empty")
	}

	path += "*"

	if nil == xml.register {
		xml.register = cedar.NewCedar()
	}
	return xml.register.Insert([]byte(path), cb)
}

func (xml *XMLReader) Parse() error {
	return xml.Iterate(noop)
}

func (xml *XMLReader) callCb(name, value, path []byte, attributes Attributes, defaultCb Iterator) error {

	if nil != xml.register {
		found := false
		var wildcardError error
		xml.register.MatchWildcard(path, 0, func(nid int, key []byte, av interface{}) {
			iterator := av.(Iterator)
			found = true
			if err := iterator(name, value, path, attributes); nil != err {
				wildcardError = err
			}
		})

		if found {
			return wildcardError
		}

	}

	return defaultCb(name, value, path, attributes)

}

func (xml *XMLReader) Iterate(cb Iterator) error {
	var curTagName, firstTagName, schemaPath []byte
	var nestedElements int
	var attributes Attributes

	defer func() {
		if len(attributes) > 0 {
			for i := 0; i < len(attributes); i++ {
				ReleaseAttribute(attributes[i])
			}
		}
	}()

	for {
		e, err := xml.r.Event()
		if err != nil {
			return err
		}

		if e.Type() == gosax.EventEOF {
			break
		}

		switch e.Type() {
		case gosax.EventStart:

			for i := 0; i < len(attributes); i++ {
				ReleaseAttribute(attributes[i])
			}
			attributes = attributes[:0]

			startTagName, bb := gosax.Name(e.Bytes)
			curTagName = append(curTagName[:0], startTagName...)
			if 0 == len(firstTagName) {
				firstTagName = append(firstTagName[:0], curTagName...)
			} else {
				if bytes.Equal(curTagName, firstTagName) {
					nestedElements++
				}
			}

			schemaPath = addSchemaPath(schemaPath, startTagName)

			for len(bb) > 0 {
				var attrReference gosax.Attribute
				attrReference, bb, err = gosax.NextAttribute(bb)
				if err != nil {
					return err
				}
				if len(attrReference.Key) == 0 {
					break
				}
				attrReference.Value, err = gosax.Unescape(attrReference.Value[1 : len(attrReference.Value)-1])
				if err != nil {
					return err
				}

				attr := AcquireAttribute()
				attr.copyFromSax(&attrReference)
				attributes = append(attributes, attr)
			}

			if bytes.Contains(e.Bytes, []byte("/>")) {
				if err = xml.callCb(curTagName, nil, schemaPath, attributes, cb); nil != err {
					return err
				}

				schemaPath = delSchemaPath(schemaPath, curTagName)
			}

		case gosax.EventText, gosax.EventCData:
			if isEmptyLine(e.Bytes) {
				continue
			}

			e.Bytes = bytes.Trim(e.Bytes, "\t\n")

			if err = xml.callCb(curTagName, e.Bytes, schemaPath, attributes, cb); nil != err {
				return err
			}

		case gosax.EventEnd:
			endTagName, _ := gosax.Name(e.Bytes)

			schemaPath = delSchemaPath(schemaPath, endTagName)

			if bytes.Equal(endTagName, firstTagName) {
				if nestedElements > 0 {
					nestedElements--
				} else {
					return nil
				}
			}

			curTagName = curTagName[:0]

		default:

		}

	}

	return nil
}

var noop = func(name, value, path []byte, attributes Attributes) error {
	return nil
}
