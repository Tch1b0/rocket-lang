package object

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type File struct {
	Filename string
	Position int64
	Handle   *os.File
}

func NewFile(name string) *File {
	return &File{Filename: name}
}

func (f *File) Type() ObjectType { return FILE_OBJ }
func (f *File) Inspect() string  { return fmt.Sprintf("<file:%s>", f.Filename) }
func (f *File) Open(mode string, perm string) error {
	if f.Filename == "!STDIN!" {
		f.Handle = os.Stdin
		return nil
	}
	if f.Filename == "!STDOUT!" {
		f.Handle = os.Stdout
		return nil
	}
	if f.Filename == "!STDERR!" {
		f.Handle = os.Stderr
		return nil
	}

	md := os.O_RDONLY

	switch mode {
	case "r":
	case "w":
		md = os.O_WRONLY
	case "wa":
		md = os.O_WRONLY | os.O_APPEND
	case "rw":
		md = os.O_RDWR
	case "rwa":
		md = os.O_RDWR | os.O_APPEND
	default:
		return fmt.Errorf("invalid file mode, got `%s`", mode)
	}

	if md != os.O_RDONLY {
		md = md | os.O_CREATE
	}

	filePerm, err := strconv.ParseUint(perm, 10, 32)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(f.Filename, md, fs.FileMode(filePerm))
	if err != nil {
		return err
	}

	f.Handle = file
	f.Position = 0

	return nil
}

func init() {
	objectMethods[FILE_OBJ] = map[string]ObjectMethod{
		"close": ObjectMethod{
			description: "Closes the file pointer. Returns always `true`.",
			returnPattern: [][]string{
				[]string{BOOLEAN_OBJ},
			},
			method: func(o Object, _ []Object) Object {
				f := o.(*File)
				f.Handle.Close()
				f.Position = -1
				return TRUE
			},
		},
		"lines": ObjectMethod{
			description: "If successfull, returns all lines of the file as array elements, otherwise `null`. Resets the position to 0 after read.",
			returnPattern: [][]string{
				[]string{ARRAY_OBJ, ERROR_OBJ},
			},
			method: func(o Object, oo []Object) Object {
				file := readFile(o, oo)
				fileString := file.(*String)
				lines := strings.Split(fileString.Value, "\n")

				result := make([]Object, len(lines))

				for i, line := range lines {
					result[i] = NewString(line)
				}

				return NewArray(result)
			},
		},
		"content": ObjectMethod{
			description: "Reads content of the file and returns it. Resets the position to 0 after read.",
			returnPattern: [][]string{
				[]string{STRING_OBJ, ERROR_OBJ},
			},
			method: readFile,
		},
		"position": ObjectMethod{
			description: "Returns the position of the current file handle. -1 if the file is closed.",
			returnPattern: [][]string{
				[]string{INTEGER_OBJ},
			},
			method: func(o Object, _ []Object) Object {
				f := o.(*File)
				return NewInteger(f.Position)
			},
		},
		"read": ObjectMethod{
			description: "Reads the given amount of bytes from the file. Sets the position to the bytes that where actually read. At the end of file EOF error is returned.",
			argPattern: [][]string{
				[]string{INTEGER_OBJ},
			},
			returnPattern: [][]string{
				[]string{STRING_OBJ, ERROR_OBJ},
			},
			method: func(o Object, args []Object) Object {
				f := o.(*File)
				bytesAmount := args[0].(*Integer).Value
				if f.Handle == nil {
					return NewError("Invalid file handle.")
				}

				buffer := make([]byte, bytesAmount)
				bytesRealRead, err := f.Handle.Read(buffer)
				f.Position += int64(bytesRealRead)

				if err != nil {
					return NewError(err)
				}

				return NewString(string(buffer))
			},
		},
		"seek": ObjectMethod{
			description: "Seek sets the offset for the next Read or Write on file to offset, interpreted according to whence. 0 means relative to the origin of the file, 1 means relative to the current offset, and 2 means relative to the end.",
			argPattern: [][]string{
				[]string{INTEGER_OBJ},
				[]string{INTEGER_OBJ},
			},
			returnPattern: [][]string{
				[]string{INTEGER_OBJ, ERROR_OBJ},
			},
			method: func(o Object, args []Object) Object {
				f := o.(*File)

				if f.Handle == nil {
					return NewError("Invalid file handle.")
				}

				seekAmount := args[0].(*Integer).Value
				seekRelative := args[1].(*Integer).Value
				newOffset, err := f.Handle.Seek(seekAmount, int(seekRelative))
				f.Position = newOffset

				if err != nil {
					return NewError(err)
				}

				return NewInteger(f.Position)
			},
		},
		"write": ObjectMethod{
			description: "Writes the given string to the file. Returns `true` on success.",
			returnPattern: [][]string{
				[]string{BOOLEAN_OBJ, ERROR_OBJ},
			},
			argPattern: [][]string{
				[]string{STRING_OBJ},
			},
			method: func(o Object, args []Object) Object {
				f := o.(*File)
				content := []byte(args[0].(*String).Value)

				if f.Handle == nil {
					return NewError("Invalid file handle.")
				}

				bytesWritten, err := f.Handle.Write(content)
				f.Position += int64(bytesWritten)

				if err != nil {
					return NewError(err)
				}

				return TRUE
			},
		},
	}
}

func (f *File) InvokeMethod(method string, env Environment, args ...Object) Object {
	return objectMethodLookup(f, method, args)
}

func readFile(o Object, _ []Object) Object {
	f := o.(*File)
	if f.Handle == nil {
		return NewError("Invalid file handle.")
	}
	if _, err := f.Handle.Seek(0, 0); err != nil {
		return NewError(err)
	}

	file, err := ioutil.ReadAll(f.Handle)
	if err != nil {
		return NewError(err)
	}

	if _, err := f.Handle.Seek(0, 0); err != nil {
		return NewError(err)
	}
	f.Position = 0
	return NewString(string(file))
}
