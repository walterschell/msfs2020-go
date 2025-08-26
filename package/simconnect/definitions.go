package simconnect

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log/slog"
	"math"
	"reflect"

	. "github.com/walterschell/msfs2020-go/package/simconnect/messages"
)

type FieldNameMissingError struct {
	Field string
}

func (e *FieldNameMissingError) Error() string {
	return "field name missing for exported field: " + e.Field
}

type FieldUnitsMissingError struct {
	Field string
}

func (e *FieldUnitsMissingError) Error() string {
	return "field units missing for exported field: " + e.Field
}

func GenerateDefinitionFor(definitionID uint32, t interface{}) ([]AddToDataDefinitionMessage, error) {
	var numExportedFields int
	for i := 0; i < reflect.TypeOf(t).NumField(); i++ {
		if reflect.TypeOf(t).Field(i).IsExported() {
			numExportedFields++
		}
	}
	if numExportedFields == 0 {
		slog.Error("No exported fields found in type", "type", reflect.TypeOf(t))
		return nil, nil // No exported fields to generate definitions for
	}
	definitions := make([]AddToDataDefinitionMessage, 0, numExportedFields)
	for i := 0; i < reflect.TypeOf(t).NumField(); i++ {
		field := reflect.TypeOf(t).Field(i)
		if !field.IsExported() {
			continue // Skip unexported fields
		}
		datumName, ok := field.Tag.Lookup("name")
		if !ok {
			return nil, &FieldNameMissingError{Field: field.Name}
		}
		datumUnits, ok := field.Tag.Lookup("units")
		if !ok {
			return nil, &FieldUnitsMissingError{Field: field.Name}
		}
		datumType := SimConnectDataTypeFloat64 // Default type, can be customized
		if field.Type.Kind() == reflect.Uint32 {
			datumType = SimConnectDataTypeInt32
		} else if field.Type.Kind() == reflect.Float32 {
			datumType = SimConnectDataTypeFloat32
		} else if field.Type.Kind() == reflect.Float64 {
			datumType = SimConnectDataTypeFloat64
		} else if field.Type.Kind() == reflect.Bool {
			datumType = SimConnectDataTypeInt32
		} else if field.Type == reflect.TypeOf(SimConnectInitPosition{}) {
			datumType = SimConnectDataTypeInitPosition
		} else {
			return nil, fmt.Errorf("unsupported field type %s for field %s", field.Type.Kind(), field.Name)
		}
		definitions = append(definitions, AddToDataDefinitionMessage{
			DefinitionID: definitionID,
			DatumName:    datumName,
			DatumUnits:   datumUnits,
			DatumType:    datumType,
			Epsilon:      0.0,       // Default epsilon, can be customized
			DatumID:      uint32(i), // Default datum ID, can be customized
		})
	}
	if len(definitions) == 0 {
		return nil, fmt.Errorf("no valid exported fields found in type %T", t)
	}
	slog.Info("Generated data definitions", "definitionID", definitionID, "type", reflect.TypeOf(t), "numDefinitions", len(definitions))
	return definitions, nil
}

func UnmarshalFromSimObjectData(v interface{}, d *RecveSimObjectDataMessage) error {
	val := reflect.ValueOf(v).Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("expected a pointer to a struct, got %T", v)
	}
	exportedFieldCount := 0
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		if field.IsExported() {
			exportedFieldCount++
		}
	}
	if exportedFieldCount != int(d.DefineCount) {
		return fmt.Errorf("expected %d exported fields, got %d", exportedFieldCount, d.DefineCount)
	}
	slog.Debug("Unmarshalling SimObjectDataMessage", "definitionID", d.DefinitionID, "objectID", d.ObjectID, "defineCount", d.DefineCount, "size", len(d.Data))
	index := 0
	for i := 0; i < val.NumField(); i++ {
		slog.Debug("Processing field", "index", i, "fieldName", val.Type().Field(i).Name, "offset", index)
		field := val.Type().Field(i)
		if !field.IsExported() {
			continue // Skip unexported fields
		}
		switch field.Type.Kind() {
		case reflect.Float64:
			if len(d.Data[index:]) < 8 {
				return fmt.Errorf("not enough data to fill field %s", field.Name)
			}
			u64 := binary.LittleEndian.Uint64(d.Data[index:])
			val.Field(i).SetFloat(math.Float64frombits(u64))
			index += 8 // Float64 is 8 bytes
		case reflect.Float32:
			if len(d.Data[index:]) < 4 {
				return fmt.Errorf("not enough data to fill field %s", field.Name)
			}
			u32 := binary.LittleEndian.Uint32(d.Data[index:])
			val.Field(i).SetFloat(float64(math.Float32frombits(u32)))
			index += 4 // Float32 is 4 bytes
		case reflect.Int32:
			if len(d.Data[index:]) < 4 {
				return fmt.Errorf("not enough data to fill field %s", field.Name)
			}
			u32 := binary.LittleEndian.Uint32(d.Data[index:])
			val.Field(i).SetInt(int64(u32))
			index += 4 // Int32 is 4 bytes

		case reflect.Uint32:
			if len(d.Data[index:]) < 4 {
				return fmt.Errorf("not enough data to fill field %s", field.Name)
			}
			u32 := binary.LittleEndian.Uint32(d.Data[index:])
			val.Field(i).SetUint(uint64(u32))
			index += 4 // Uint32 is 4 bytes

		default:
			return fmt.Errorf("unsupported field type %s for field %s", field.Type.Kind(), field.Name)

		}
	}
	return nil
}

func MarshalToSimObjectData(v interface{}) ([]byte, error) {
	slog.Info("Marshalling SimObjectData", "type", reflect.TypeOf(v))
	val := reflect.ValueOf(v)
	data := make([]byte, 0)
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		if !field.IsExported() {
			continue // Skip unexported fields
		}
		switch field.Type.Kind() {
		case reflect.Float64:
			u64 := math.Float64bits(val.Field(i).Float())
			data = binary.LittleEndian.AppendUint64(data, u64)
		case reflect.Float32:
			u32 := math.Float32bits(float32(val.Field(i).Float()))
			data = binary.LittleEndian.AppendUint32(data, u32)
		case reflect.Int32:
			u32 := uint32(val.Field(i).Int())
			data = binary.LittleEndian.AppendUint32(data, u32)
		case reflect.Uint32:
			// Directly append uint32 value
			u32 := val.Field(i).Uint()
			if u32 > math.MaxUint32 {
				return nil, fmt.Errorf("value for field %s exceeds uint32 range", field.Name)
			}
			data = binary.LittleEndian.AppendUint32(data, uint32(u32))
		case reflect.Bool:
			// Convert bool to uint32 (0 or 1)
			var u32 uint32
			if val.Field(i).Bool() {
				u32 = 1
			} else {
				u32 = 0
			}
			data = binary.LittleEndian.AppendUint32(data, u32)
		case reflect.Pointer:
			if val.Field(i).IsNil() {
				return nil, fmt.Errorf("nil pointer for field %s", field.Name)
			}
			elem := val.Field(i).Elem()
			if elem.Kind() != reflect.Struct {
				return nil, fmt.Errorf("unsupported pointer type for field %s, expected pointer to struct", field.Name)
			}
			if elem.Type() == reflect.TypeOf(SimConnectInitPosition{}) {
				pos := elem.Interface().(SimConnectInitPosition)
				data = append(data, pos.MarshalBinary()...)
			} else {
				return nil, fmt.Errorf("unsupported pointer struct type %s for field %s", elem.Type(), field.Name)
			}
		case reflect.Struct:
			if field.Type == reflect.TypeOf(SimConnectInitPosition{}) {
				pos := val.Field(i).Interface().(SimConnectInitPosition)
				data = append(data, pos.MarshalBinary()...)
			} else {
				return nil, fmt.Errorf("unsupported struct type %s for field %s", field.Type, field.Name)
			}
		default:
			return nil, fmt.Errorf("unsupported field type %s for field %s", field.Type.Kind(), field.Name)
		}
	}
	return data, nil
}

type TeleportRequest struct {
	Latitude  float64 `name:"PLANE LATITUDE" units:"degrees"`
	Longitude float64 `name:"PLANE LONGITUDE" units:"degrees"`
}

type TeleportRequestWithAltitude struct {
	Latitude  float64 `name:"PLANE LATITUDE" units:"degrees"`
	Longitude float64 `name:"PLANE LONGITUDE" units:"degrees"`
	Altitude  float64 `name:"PLANE ALTITUDE" units:"feet"`
}

type ColdStartTeleportRequest struct {
	InitPos       SimConnectInitPosition `name:"INITIAL POSITION" units:""`
	MasterBattery bool                   `name:"ELECTRICAL MASTER BATTERY" units:"bool"`
	//EngineControlSelect   uint32                 `name:"ENGINE CONTROL SELECT" units:"flags"`
	GeneralEngCombustionn bool `name:"GENERAL ENG COMBUSTION:1" units:"bool"`
	GeneratorActive       bool `name:"GENERAL ENG GENERATOR ACTIVE:1" units:"bool"`
}

func NewColdStartTeleportRequest(lat, lon float64) ColdStartTeleportRequest {
	return ColdStartTeleportRequest{
		InitPos: SimConnectInitPosition{
			Latitude:  lat,
			Longitude: lon,
			Altitude:  0.0,
			Pitch:     0.0,
			Bank:      0.0,
			Heading:   0.0,
			OnGround:  1,
			Airspeed:  0,
		},
		MasterBattery: false,
		//EngineControlSelect:   0b1111, // All engines selected
		GeneralEngCombustionn: false,
		GeneratorActive:       false,
	}
}

/*
double  Latitude;
double  Longitude;
double  Altitude;
double  Pitch;
double  Bank;
double  Heading;
DWORD  OnGround;
DWORD  Airspeed;
*/
type SimConnectInitPosition struct {
	Latitude  float64 // degrees
	Longitude float64 // degrees
	Altitude  float64 // feet
	Pitch     float64 // degrees
	Bank      float64 // degrees
	Heading   float64 // degrees
	OnGround  uint32  // bool as uint32
	Airspeed  uint32  // knots
}

func (p *SimConnectInitPosition) MarshalBinary() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, p.Latitude)
	binary.Write(buf, binary.LittleEndian, p.Longitude)
	binary.Write(buf, binary.LittleEndian, p.Altitude)
	binary.Write(buf, binary.LittleEndian, p.Pitch)
	binary.Write(buf, binary.LittleEndian, p.Bank)
	binary.Write(buf, binary.LittleEndian, p.Heading)
	binary.Write(buf, binary.LittleEndian, p.OnGround)
	binary.Write(buf, binary.LittleEndian, p.Airspeed)
	return buf.Bytes()
}
