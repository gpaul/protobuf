// Copyright (c) 2013, Vastech SA (PTY) LTD. All rights reserved.
// http://code.google.com/p/gogoprotobuf
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package unmarshal

import (
	"code.google.com/p/gogoprotobuf/gogoproto"
	"code.google.com/p/gogoprotobuf/proto"
	descriptor "code.google.com/p/gogoprotobuf/protoc-gen-gogo/descriptor"
	"code.google.com/p/gogoprotobuf/protoc-gen-gogo/generator"
	//"path"
	"strconv"
	"strings"
)

type unmarshal struct {
	*generator.Generator
	generator.PluginImports
	atleastOne bool
	ioPkg      generator.Single
	localNum   string
}

func NewUnmarshal() *unmarshal {
	return &unmarshal{}
}

func (p *unmarshal) Name() string {
	return "unmarshal"
}

func (p *unmarshal) Init(g *generator.Generator) {
	p.Generator = g
}

func wireToType(wire string) int {
	switch wire {
	case "fixed64":
		return proto.WireFixed64
	case "fixed32":
		return proto.WireFixed32
	case "varint":
		return proto.WireVarint
	case "bytes":
		return proto.WireBytes
	case "group":
		return proto.WireBytes
	case "zigzag32":
		return proto.WireVarint
	case "zigzag64":
		return proto.WireVarint
	}
	panic("unreachable")
}

func (p *unmarshal) decodeVarint(varName string, typName string) {
	p.P(`for shift := uint(0); ; shift += 7 {`)
	p.In()
	p.P(`if index >= l {`)
	p.In()
	p.P(`return `, p.ioPkg.Use(), `.ErrUnexpectedEOF`)
	p.Out()
	p.P(`}`)
	p.P(`b := data[index]`)
	p.P(`index++`)
	p.P(varName, ` |= (`, typName, `(b) & 0x7F) << shift`)
	p.P(`if b < 0x80 {`)
	p.In()
	p.P(`break`)
	p.Out()
	p.P(`}`)
	p.Out()
	p.P(`}`)
}

func (p *unmarshal) decodeFixed32(varName string, typeName string) {
	p.P(`i := index + 4`)
	p.P(`if i > l {`)
	p.In()
	p.P(`return `, p.ioPkg.Use(), `.ErrUnexpectedEOF`)
	p.Out()
	p.P(`}`)
	p.P(`index = i`)
	p.P(varName, ` = `, typeName, `(data[i-4])`)
	p.P(varName, ` |= `, typeName, `(data[i-3]) << 8`)
	p.P(varName, ` |= `, typeName, `(data[i-2]) << 16`)
	p.P(varName, ` |= `, typeName, `(data[i-1]) << 24`)
}

func (p *unmarshal) decodeFixed64(varName string, typeName string) {
	p.P(`i := index + 8`)
	p.P(`if i > l {`)
	p.In()
	p.P(`return `, p.ioPkg.Use(), `.ErrUnexpectedEOF`)
	p.Out()
	p.P(`}`)
	p.P(`index = i`)
	p.P(varName, ` = `, typeName, `(data[i-8])`)
	p.P(varName, ` |= `, typeName, `(data[i-7]) << 8`)
	p.P(varName, ` |= `, typeName, `(data[i-6]) << 16`)
	p.P(varName, ` |= `, typeName, `(data[i-5]) << 24`)
	p.P(varName, ` |= `, typeName, `(data[i-4]) << 32`)
	p.P(varName, ` |= `, typeName, `(data[i-3]) << 40`)
	p.P(varName, ` |= `, typeName, `(data[i-2]) << 48`)
	p.P(varName, ` |= `, typeName, `(data[i-1]) << 56`)
}

func (p *unmarshal) Generate(file *generator.FileDescriptor) {
	p.PluginImports = generator.NewPluginImports(p.Generator)
	p.atleastOne = false

	//fileName := path.Base(file.GetName())
	//packageName := file.PackageName()
	p.ioPkg = p.NewImport("io")
	mathPkg := p.NewImport("math")

	for _, message := range file.Messages() {
		if !gogoproto.IsUnmarshaler(file.FileDescriptorProto, message.DescriptorProto) {
			continue
		}
		p.atleastOne = true

		ccTypeName := generator.CamelCaseSlice(message.TypeName())
		p.P(`func (m *`, ccTypeName, `) Unmarshal(data []byte) error {`)
		p.In()
		p.P(`l := len(data)`)
		p.P(`index := 0`)
		p.P(`for index < l {`)
		p.In()
		p.P(`var wire uint64`)
		p.decodeVarint("wire", "uint64")
		p.P(`fieldNum := int32(wire >> 3)`)
		if len(message.Field) > 0 {
			p.P(`wireType := int(wire & 0x7)`)
		}
		p.P(`switch fieldNum {`)
		p.In()
		for _, field := range message.Field {
			fieldname := generator.CamelCase(*field.Name)
			repeated := field.IsRepeated()
			nullable := gogoproto.IsNullable(field)
			packed := field.IsPacked()
			p.P(`case `, strconv.Itoa(int(field.GetNumber())), `:`)
			p.In()
			_, wire := p.GoType(message, field)
			wireType := wireToType(wire)
			if !packed {
				p.P(`if wireType != `, strconv.Itoa(wireType), `{`)
				p.In()
				p.P(`return proto.ErrWrongType`)
				p.Out()
				p.P(`}`)
			} else {
				p.P(`if wireType != `, strconv.Itoa(proto.WireBytes), `{`)
				p.In()
				p.P(`return proto.ErrWrongType`)
				p.Out()
				p.P(`}`)
			}
			if packed {
				p.P(`var packedLen int`)
				p.decodeVarint("packedLen", "int")
				p.P(`postIndex := index + packedLen`)
				p.P(`if postIndex > l {`)
				p.In()
				p.P(`return `, p.ioPkg.Use(), `.ErrUnexpectedEOF`)
				p.Out()
				p.P(`}`)
				p.P(`for index < postIndex {`)
				p.In()
			}
			switch *field.Type {
			case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
				if repeated {
					p.P(`var v uint64`)
					p.decodeFixed64("v", "uint64")
					p.P(`v2 := `, mathPkg.Use(), `.Float64frombits(v)`)
					p.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v2)`)
				} else if nullable {
					p.P(`var v uint64`)
					p.decodeFixed64("v", "uint64")
					p.P(`v2 := `, mathPkg.Use(), `.Float64frombits(v)`)
					p.P(`m.`, fieldname, ` = &v2`)
				} else {
					p.P(`var v uint64`)
					p.decodeFixed64("v", "uint64")
					p.P(`m.`, fieldname, ` = `, mathPkg.Use(), `.Float64frombits(v)`)
				}
			case descriptor.FieldDescriptorProto_TYPE_FLOAT:
				if repeated {
					p.P(`var v uint32`)
					p.decodeFixed32("v", "uint32")
					p.P(`v2 := `, mathPkg.Use(), `.Float32frombits(v)`)
					p.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v2)`)
				} else if nullable {
					p.P(`var v uint32`)
					p.decodeFixed32("v", "uint32")
					p.P(`v2 := `, mathPkg.Use(), `.Float32frombits(v)`)
					p.P(`m.`, fieldname, ` = &v2`)
				} else {
					p.P(`var v uint32`)
					p.decodeFixed32("v", "uint32")
					p.P(`m.`, fieldname, ` = `, mathPkg.Use(), `.Float32frombits(v)`)
				}
			case descriptor.FieldDescriptorProto_TYPE_INT64:
				if repeated {
					p.P(`var v int64`)
					p.decodeVarint("v", "int64")
					p.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v)`)
				} else if nullable {
					p.P(`var v int64`)
					p.decodeVarint("v", "int64")
					p.P(`m.`, fieldname, ` = &v`)
				} else {
					p.decodeVarint("m."+fieldname, "int64")
				}
			case descriptor.FieldDescriptorProto_TYPE_UINT64:
				if repeated {
					p.P(`var v uint64`)
					p.decodeVarint("v", "uint64")
					p.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v)`)
				} else if nullable {
					p.P(`var v uint64`)
					p.decodeVarint("v", "uint64")
					p.P(`m.`, fieldname, ` = &v`)
				} else {
					p.decodeVarint("m."+fieldname, "uint64")
				}
			case descriptor.FieldDescriptorProto_TYPE_INT32:
				if repeated {
					p.P(`var v int32`)
					p.decodeVarint("v", "int32")
					p.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v)`)
				} else if nullable {
					p.P(`var v int32`)
					p.decodeVarint("v", "int32")
					p.P(`m.`, fieldname, ` = &v`)
				} else {
					p.decodeVarint("m."+fieldname, "int32")
				}
			case descriptor.FieldDescriptorProto_TYPE_FIXED64:
				if repeated {
					p.P(`var v uint64`)
					p.decodeFixed64("v", "uint64")
					p.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v)`)
				} else if nullable {
					p.P(`var v uint64`)
					p.decodeFixed64("v", "uint64")
					p.P(`m.`, fieldname, ` = &v`)
				} else {
					p.decodeFixed64("m."+fieldname, "uint64")
				}
			case descriptor.FieldDescriptorProto_TYPE_FIXED32:
				if repeated {
					p.P(`var v uint32`)
					p.decodeFixed32("v", "uint32")
					p.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v)`)
				} else if nullable {
					p.P(`var v uint32`)
					p.decodeFixed32("v", "uint32")
					p.P(`m.`, fieldname, ` = &v`)
				} else {
					p.decodeFixed32("m."+fieldname, "uint32")
				}
			case descriptor.FieldDescriptorProto_TYPE_BOOL:
				if repeated {
					p.P(`var v int`)
					p.decodeVarint("v", "int")
					p.P(`m.`, fieldname, ` = append(m.`, fieldname, `, bool(v != 0))`)
				} else if nullable {
					p.P(`var v int`)
					p.decodeVarint("v", "int")
					p.P(`b := bool(v != 0)`)
					p.P(`m.`, fieldname, ` = &b`)
				} else {
					p.P(`var v int`)
					p.decodeVarint("v", "int")
					p.P(`m.`, fieldname, ` = bool(v != 0)`)
				}
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				p.P(`var stringLen uint64`)
				p.decodeVarint("stringLen", "uint64")
				p.P(`postIndex := index + int(stringLen)`)
				p.P(`if postIndex > l {`)
				p.In()
				p.P(`return `, p.ioPkg.Use(), `.ErrUnexpectedEOF`)
				p.Out()
				p.P(`}`)
				if repeated {
					p.P(`m.`, fieldname, ` = append(m.`, fieldname, `, string(data[index:postIndex]))`)
				} else if nullable {
					p.P(`s := string(data[index:postIndex])`)
					p.P(`m.`, fieldname, ` = &s`)
				} else {
					p.P(`m.`, fieldname, ` = string(data[index:postIndex])`)
				}
				p.P(`index = postIndex`)
			case descriptor.FieldDescriptorProto_TYPE_GROUP:
				panic("groups are not supported")
			case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
				desc := p.ObjectNamed(field.GetTypeName())
				msgname := p.TypeName(desc)
				msgnames := strings.Split(msgname, ".")
				typeName := msgnames[len(msgnames)-1]
				if gogoproto.IsEmbed(field) {
					fieldname = typeName
				}
				p.P(`var msglen int`)
				p.decodeVarint("msglen", "int")
				p.P(`postIndex := index + msglen`)
				p.P(`if postIndex > l {`)
				p.In()
				p.P(`return `, p.ioPkg.Use(), `.ErrUnexpectedEOF`)
				p.Out()
				p.P(`}`)
				if repeated {
					if nullable {
						p.P(`m.`, fieldname, ` = append(m.`, fieldname, `, &`, msgname, `{})`)
					} else {
						p.P(`m.`, fieldname, ` = append(m.`, fieldname, `, `, msgname, `{})`)
					}
					p.P(`m.`, fieldname, `[len(m.`, fieldname, `)-1].Unmarshal(data[index:postIndex])`)
				} else if nullable {
					p.P(`m.`, fieldname, ` = &`, msgname, `{}`)
					p.P(`if err := m.`, fieldname, `.Unmarshal(data[index:postIndex]); err != nil {`)
					p.In()
					p.P(`return err`)
					p.Out()
					p.P(`}`)
				} else {
					p.P(`if err := m.`, fieldname, `.Unmarshal(data[index:postIndex]); err != nil {`)
					p.In()
					p.P(`return err`)
					p.Out()
					p.P(`}`)
				}
				p.P(`index = postIndex`)
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				p.P(`var byteLen int`)
				p.decodeVarint("byteLen", "int")
				p.P(`postIndex := index + byteLen`)
				p.P(`if postIndex > l {`)
				p.In()
				p.P(`return `, p.ioPkg.Use(), `.ErrUnexpectedEOF`)
				p.Out()
				p.P(`}`)
				if !gogoproto.IsCustomType(field) {
					if repeated {
						p.P(`m.`, fieldname, ` = append(m.`, fieldname, `, data[index:postIndex])`)
					} else if nullable {
						p.P(`m.`, fieldname, ` = data[index:postIndex]`)
					} else {
						p.P(`m.`, fieldname, ` = data[index:postIndex]`)
					}
				} else {
					packageName, _, err := generator.GetCustomType(field)
					if err != nil {
						panic(err)
					}
					if repeated {
						p.P(`m.`, fieldname, ` = append(m.`, fieldname, `, `, packageName, `{})`)
						p.P(`m.`, fieldname, `[len(m.`, fieldname, `)-1].Unmarshal(data[index:postIndex])`)
					} else if nullable {
						p.P(`m.`, fieldname, ` = &`, packageName, `{}`)
						p.P(`if err := m.`, fieldname, `.Unmarshal(data[index:postIndex]); err != nil {`)
						p.In()
						p.P(`return err`)
						p.Out()
						p.P(`}`)
					} else {
						p.P(`if err := m.`, fieldname, `.Unmarshal(data[index:postIndex]); err != nil {`)
						p.In()
						p.P(`return err`)
						p.Out()
						p.P(`}`)
					}
				}
				p.P(`index = postIndex`)
			case descriptor.FieldDescriptorProto_TYPE_UINT32:
				if repeated {
					p.P(`var v uint32`)
					p.decodeVarint("v", "uint32")
					p.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v)`)
				} else if nullable {
					p.P(`var v uint32`)
					p.decodeVarint("v", "uint32")
					p.P(`m.`, fieldname, ` = &v`)
				} else {
					p.decodeVarint("m."+fieldname, "uint32")
				}
			case descriptor.FieldDescriptorProto_TYPE_ENUM:
				typName := p.TypeName(p.ObjectNamed(field.GetTypeName()))
				if repeated {
					p.P(`var v `, typName)
					p.decodeVarint("v", typName)
					p.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v)`)
				} else if nullable {
					p.P(`var v `, typName)
					p.decodeVarint("v", typName)
					p.P(`m.`, fieldname, ` = &v`)
				} else {
					p.decodeVarint("m."+fieldname, typName)
				}
			case descriptor.FieldDescriptorProto_TYPE_SFIXED32:
				if repeated {
					p.P(`var v int32`)
					p.decodeFixed32("v", "int32")
					p.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v)`)
				} else if nullable {
					p.P(`var v int32`)
					p.decodeFixed32("v", "int32")
					p.P(`m.`, fieldname, ` = &v`)
				} else {
					p.decodeFixed32("m."+fieldname, "int32")
				}
			case descriptor.FieldDescriptorProto_TYPE_SFIXED64:
				if repeated {
					p.P(`var v int64`)
					p.decodeFixed64("v", "int64")
					p.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v)`)
				} else if nullable {
					p.P(`var v int64`)
					p.decodeFixed64("v", "int64")
					p.P(`m.`, fieldname, ` = &v`)
				} else {
					p.decodeFixed64("m."+fieldname, "int64")
				}
			case descriptor.FieldDescriptorProto_TYPE_SINT32:
				p.P(`var v int32`)
				p.decodeVarint("v", "int32")
				p.P(`v = int32((uint32(v) >> 1) ^ uint32(((v&1)<<31)>>31))`)
				if repeated {
					p.P(`m.`, fieldname, ` = append(m.`, fieldname, `, v)`)
				} else if nullable {
					p.P(`m.`, fieldname, ` = &v`)
				} else {
					p.P(`m.`, fieldname, ` = v`)
				}
			case descriptor.FieldDescriptorProto_TYPE_SINT64:
				p.P(`var v uint64`)
				p.decodeVarint("v", "uint64")
				p.P(`v = (v >> 1) ^ uint64((int64(v&1)<<63)>>63)`)
				if repeated {
					p.P(`m.`, fieldname, ` = append(m.`, fieldname, `, int64(v))`)
				} else if nullable {
					p.P(`v2 := int64(v)`)
					p.P(`m.`, fieldname, ` = &v2`)
				} else {
					p.P(`m.`, fieldname, ` = int64(v)`)
				}
			default:
				panic("not implemented")
			}
			if packed {
				p.Out()
				p.P(`}`)
			}
		}
		p.Out()
		p.P(`default:`)
		p.In()

		p.P(`key := uint32(fieldNum)<<3 | uint32(wireType)`)
		p.P(`for key > 127 {`)
		p.In()
		p.P(`m.XXX_unrecognized = append(m.XXX_unrecognized, 0x80|uint8(key&0x7F))`)
		p.P(`key >>= 7`)
		p.Out()
		p.P(`}`)
		p.P(`m.XXX_unrecognized = append(m.XXX_unrecognized, uint8(key))`)

		p.P(`switch wireType {`)

		p.P(`case 0:`)
		p.In()
		p.P(`for {`)
		p.In()
		p.P(`if index >= l {`)
		p.In()
		p.P(`return `, p.ioPkg.Use(), `.ErrUnexpectedEOF`)
		p.Out()
		p.P(`}`)
		p.P(`m.XXX_unrecognized = append(m.XXX_unrecognized, data[index])`)
		p.P(`index++`)
		p.P(`if data[index-1] < 0x80 {`)
		p.In()
		p.P(`break`)
		p.Out()
		p.P(`}`)
		p.Out()
		p.P(`}`)
		p.Out()

		p.P(`case 1:`)
		p.In()
		p.P(`m.XXX_unrecognized = append(m.XXX_unrecognized, data[index:index+8]...)`)
		p.P(`index+=8`)
		p.Out()

		p.P(`case 2:`)
		p.In()

		p.P(`var length int`)
		p.P(`for shift := uint(0); ; shift += 7 {`)
		p.In()
		p.P(`if index >= l {`)
		p.In()
		p.P(`return `, p.ioPkg.Use(), `.ErrUnexpectedEOF`)
		p.Out()
		p.P(`}`)
		p.P(`b := data[index]`)
		p.P(`m.XXX_unrecognized = append(m.XXX_unrecognized, b)`)
		p.P(`index++`)
		p.P(`length |= (int(b) & 0x7F) << shift`)
		p.P(`if b < 0x80 {`)
		p.In()
		p.P(`break`)
		p.Out()
		p.P(`}`)
		p.Out()
		p.P(`}`)

		p.P(`m.XXX_unrecognized = append(m.XXX_unrecognized, data[index:index+length]...)`)
		p.P(`index+=length`)

		p.Out()

		p.P(`case 3:`)
		p.In()
		p.Out()

		p.P(`case 4:`)
		p.In()
		p.Out()

		p.P(`case 5:`)
		p.In()
		p.P(`m.XXX_unrecognized = append(m.XXX_unrecognized, data[index:index+4]...)`)
		p.P(`index+=4`)
		p.Out()

		p.P(`default:`)
		p.In()
		p.P(`panic("not implemented")`)
		p.Out()
		p.P(`}`)
		p.Out()
		p.Out()
		p.P(`}`)
		p.Out()
		p.P(`}`)
		p.P(`return nil`)
		p.Out()
		p.P(`}`)
	}

	if !p.atleastOne {
		return
	}

}

func init() {
	generator.RegisterPlugin(NewUnmarshal())
}