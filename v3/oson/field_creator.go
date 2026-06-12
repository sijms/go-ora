package oson

type ObjectFieldCreator struct{}

//func (obj *ObjectFieldCreator) CreateField(input interface{}, header *Header) (Field, error) {
//	if temp, ok := input.(map[string]interface{}); ok {
//		return NewObjectField(temp, header)
//	}
//	return nil, errors.New("invalid input for ObjectField")
//}
//
//type NumberFieldCreator struct {
//	encoder NumberEncoder
//	decoder NumberDecoder
//}
//
//func (num *NumberFieldCreator) CreateField(input interface{}, header *Header) (Field, error) {
//	//data, err := num.encoder.EncodeNumber(input)
//	//if err != nil {
//	//	return nil, err
//	//}
//	return &NumberField{value: *data}, nil
//}
//
//type BinaryDoubleFieldCreator struct {
//	encoder BinaryDoubleEncoder
//	decoder BinaryDoubleDecoder
//}
//
//func (num *BinaryDoubleFieldCreator) CreateField(input interface{}, header *Header) (Field, error) {
//	if temp, ok := input.(float64); ok {
//		data, err := num.encoder.EncodeBinaryDouble(temp)
//		if err != nil {
//			return nil, err
//		}
//		return &BinaryDoubleField{data: data, decoder: num.decoder}, nil
//	}
//	if temp, ok := input.(float32); ok {
//		data, err := num.encoder.EncodeBinaryDouble(float64(temp))
//		if err != nil {
//			return nil, err
//		}
//		return &BinaryDoubleField{data: data, decoder: num.decoder}, nil
//	}
//	return nil, errors.New("invalid input for BinaryDoubleField")
//}
//
//type BinaryFloatFieldCreator struct {
//	encoder BinaryFloatEncoder
//	decoder BinaryFloatDecoder
//}
//
//func (num *BinaryFloatFieldCreator) CreateField(input interface{}, header *Header) (Field, error) {
//	if temp, ok := input.(float32); ok {
//		data, err := num.encoder.EncodeBinaryFloat(temp)
//		if err != nil {
//			return nil, err
//		}
//		return &BinaryFloatField{data: data, decoder: num.decoder}, nil
//	}
//	if temp, ok := input.(float64); ok {
//		data, err := num.encoder.EncodeBinaryFloat(float32(temp))
//		if err != nil {
//			return nil, err
//		}
//		return &BinaryFloatField{data: data, decoder: num.decoder}, nil
//	}
//	return nil, errors.New("invalid input for BinaryFloatField")
//}
