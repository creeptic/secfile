# Secret file
Minimal wrapper around `os.File` that encrypts/decrypts all data read/written 
to it

### Details
Function `Open()` conforms to the `os.OpenFile()` interface, taking filename, 
flags and permissions as well as special `key` parameter which is used to 
construct steam cipher. Implementation expects 256-bit keys for use in AES-256
to construct AES-CTR cipher with random nonce. Nonce itself is stored at the 
beginning of file and is used to reconstruct cipher each time file is opened 
for reading or appending.

### `io` interfaces
Type `Secfile` implements following interfaces:
- io.Reader
- io.Writer
- io.Closer

### TODO
`io.Seeker` should be implemented as well to complete *Go*'s `io` interface
suite; this will probably also allow to use `os.O_RDWR`, if rolling key stream
back and forth can be implemented without too much overhead.
