# go-copier

I am a struct copier. I copy everything I can and copy more than https://github.com/jinzhu/copier. I also have support for 
protobuf wrappers.

* Copy from field to field with same name
* Copy from slice to slice
* Copy from pointer to non-pointer of same or compatible type
* Copy from map to map
* Ignore a field with a tag
* Map a field with different names with a tag
* Deep Copy
* Copy from encoding.TextMarshaler to string
* Copy from string to encoding.TextUnmarshaler
* Support type aliases and string enums
* Override Marshal behavior with CopierMarshaler
* Override UnMarshal behavior with CopierUnmarshal
* Copy between unsigned int and *int types (e.g. uint16 to uint32)
* Copy between signed int and *int types (e.g. int16 to int32) 
* Copy timestamppb.Timestamp to time.Time and *time.Time
* Copy time.Time and *time.Time to timestamppb.Timestamp
* Copy durationpb.Duration to time.Duration and *time.Duration
* Copy time.Duration and *time.Duration to durationpb.Duration
* Copy from wrapperspb.StringValue to string and *string
* Copy from string and *string to wrapperspb.StringValue
* Copy from wrapperspb.BoolValue to bool and *bool
* Copy from bool and *bool to wrapperspb.BoolValue
* Copy from wrapperspb.Int32Value to any signed integer or signed integer pointer 
* Copy from any signed integer or signed integer pointer to wrapperspb.Int32Value
* Copy from wrapperspb.Int64Value to any signed integer or signed integer pointer
* Copy from any signed integer or signed integer pointer to wrapperspb.Int64Value
* Copy from wrapperspb.UInt32Value to any unsigned integer or unsigned integer pointer
* Copy from any unsigned integer or unsigned integer pointer to wrapperspb.UInt32Value
* Copy from wrapperspb.UInt64Value to any unsigned integer or unsigned integer pointer
* Copy from any unsigned integer or unsigned integer pointer to wrapperspb.UInt64Value
* Copy from wrapperspb.FloatValue to any float or float pointer
* Copy from any float or float pointer to wrapperspb.FloatValue
* Copy from wrapperspb.DoubleValue to any float or float pointer
* Copy from any float or float pointer to wrapperspb.DoubleValue
* Copy from wrapperspb.BytesValue to []byte
* Copy from []byte to wrapperspb.BytesValue

## Usage

```
err := copier.Copy(dstStructPtr, srcStructPtr)
```

See copier_test.go for more

### Unit Tests
```
make test
```

* Linter

Requires golangci-lint.
```
brew install golangci/tap/golangci-lint
make golint
```

* Security Linter

```
make gosec
```

## Contributing

Happy to accept PRs.

# Author

**davidwartell**

* <http://github.com/davidwartell>
* <http://linkedin.com/in/wartell>

## License

Released under the [Apache License](https://github.com/davidwartell/go-copier/blob/master/LICENSE).

