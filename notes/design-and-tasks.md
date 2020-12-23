# Design & Tasks
Just my rambling & thoughts on Kagi's design & listing tasks in a somewhat organized fashion.  
Most of this is liable to (and will likely) change. Think of this as discovery process document.
## File System
### Design:
- Kagi is not an in-memory storage, and therefore uses a local file for presistence.  
- Since this is a key-value db, JSON format seems like the intuitive fit.  
- Not utilizing a tree structure for cursor (such as B+Tree), and instead simply accessing JSON data by key.
- For compression, Kagi uses [zstd](https://github.com/DataDog/zstd)
### Tasks
[ ] Open file
[ ] Close file
[ ] Write to a file
[ ] Read from the file
[ ] On db close, compress the file using a compression package
[ ] On db open, uncompress the file

## Locking
### Design:
- Kagi locks file during operations.
- There is only one lock named `dblock`.
### Tasks:
[ ] Lock `dblock` before writing
[ ] Lock `dblock` before reading

## Set/Get
### Design:
- As is common for key-value db, kagi utilizes `set` & `get` for storing & accessing data respectively.
- Spaces
  * When storing data with `set`, a `space` can also be set.
  * A `space` is a grouping of keys that are loosely related.
  * A `space` must already exist before adding a key/value to it
  * Spaces are stored as key/value pair, where key is the name of the space, and value is the list of keys in the space.
### Tasks:
[ ] Set stores given data using given key
  [ ] If key does not exist, create key
  [ ] If key exists, update value
[ ] Get retrieves data given a key
  [ ] If key exists, return value
  [ ] If key does not exist, return null
[ ] Delete a key/value pair
[ ] Create a space
[ ] Delete a space

## Error Handling
### Design:
- Be concise & not overly verbose, this isn't a complex db & therefore the errors neednt be complex themselves.
- Errors are thrown only when a db transaction cannot be carried out
### Tasks:
[ ] Define set of errors, for starters:
  - Key does not exist
  - Space does not exist
  - Error writing to file (if something happens to file while db is open)

## Logging
### Design:
- Nothing special here, just your average regular logger.
- Logs are saved in the db in a separate space named `kagi-logs`
- For each log, the key is the unix timestamp, and the value is the logged error/data
### Tasks:
[ ] On db open start a logging session (clearing previous logs? not sure if this is the right approach)
[ ] During this session log all db changes, warning, & errors

## Where (idea)
### Design
- Where is a method of finding all key/value pairs in the db that fulfill an equality given a user function.
- Goal is to be (somewhat) similar to the use of Where in SQL.
### Tasks (N/A)
