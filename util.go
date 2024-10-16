package fastxml

import (
	"bytes"
)

func addSchemaPath(schema, element []byte) []byte {
	if len(schema) > 0 {
		schema = append(schema, '.')
	}
	schema = append(schema, element...)
	return schema
}

func delSchemaPath(schema, element []byte) []byte {
	ss := len(schema)
	es := len(element)
	if es+1 >= ss {
		return schema[:0]
	}

	lastElementInPath := schema[(ss - es):]
	if !bytes.Equal(lastElementInPath, element) {
		return schema
	}

	return schema[:ss-(es+1)]
}

func isEmptyLine(str []byte) bool {
	size := len(str)
	if size <= 2 {
		return true
	}

	for i := 0; i < size; i++ {
		b := str[i]
		if b != '\n' && b != '\t' && b != ' ' {
			return false
		}
	}

	return true
}
