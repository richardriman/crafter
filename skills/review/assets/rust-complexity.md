## Rust Code Complexity Rules

### Overview
Rust's ownership system, pattern matching, and error handling with Result/Option add unique complexity considerations. The borrow checker enforces memory safety but can lead to code patterns that inflate traditional metrics. The following rules account for Rust-specific patterns.

### Standard Functions
All helper functions, utility functions, and non-routing functions must adhere to these limits:

| Metric | Limit | Notes |
|--------|-------|-------|
| Cyclomatic complexity | ≤ 12 | Pattern matching and Result/Option increase branch count |
| Cognitive complexity | ≤ 12 | Nested match/if let expressions reduce readability |
| Function length | < 50 lines | Excluding comments and blank lines |
| Match arms | ≤ 10 | Per single match expression |
| Nesting depth | ≤ 4 levels | Nested match, if, loops, closures |
| Parameters | ≤ 5 | Use struct for more parameters |
| Chained method calls | ≤ 7 | Split longer iterator chains into named steps |

### Error Handling Functions
Functions with extensive error handling (multiple `?` operators, match on Result/Option) have adjusted limits:

| Metric | Limit | Notes |
|--------|-------|-------|
| Cyclomatic complexity | ≤ 20 | Each `?` and match arm adds to complexity |
| Function length | < 60 lines | Error handling adds boilerplate |

### Match Expression Guidelines
Match expressions act as routers similar to Elm's case expressions:

| Metric | Limit | Notes |
|--------|-------|-------|
| Match arms in dispatch functions | No limit | Function acts as message/event router |
| **Single arm length** | ≤ 5 lines | Inline code per arm |
| **Arm complexity** | ≤ 1 expression | Extract to helper function if more complex |

#### Correct Pattern
```rust
fn handle_event(event: Event, state: &mut State) -> Result<(), Error> {
    match event {
        Event::UserLoggedIn(user) => handle_login(user, state),
        Event::UserLoggedOut => handle_logout(state),
        Event::DataReceived(data) => handle_data(data, state),
        Event::ErrorOccurred(err) => handle_error(err, state),
    }
}

fn handle_login(user: User, state: &mut State) -> Result<(), Error> {
    // Complex logic here - max 50 lines
    ...
}
```

#### Incorrect Pattern
```rust
fn handle_event(event: Event, state: &mut State) -> Result<(), Error> {
    match event {
        Event::UserLoggedIn(user) => {
            let validated_user = validate_user(&user)?;
            state.current_user = Some(validated_user.clone());
            state.is_authenticated = true;
            log::info!("User {} logged in", validated_user.name);
            send_notification(&validated_user)?;
            update_last_login(&validated_user)?;
            Ok(())
        }
        // ... more arms with inline logic
    }
}
```

### Module-Level Limits

| Metric | Limit | Action when exceeded |
|--------|-------|----------------------|
| Module length | < 1000 lines | Split into submodules |
| Public functions per module | ≤ 15 | Module may have too many responsibilities |
| Impl blocks per struct | ≤ 3 | Consider trait extraction |
| Methods per impl block | ≤ 12 | Split into multiple impl blocks or traits (exception: GraphQL resolver impl blocks) |

### Narrowing Function Arguments
Functions should accept the narrowest possible type that satisfies their requirements.

#### Principles
- Accept references (`&T`) instead of owned values when ownership is not needed
- Use trait bounds instead of concrete types when possible
- Accept slices (`&[T]`) instead of `Vec<T>` when only reading
- Accept `&str` instead of `String` when only reading
- Pass only the data the function actually needs, not entire structs

#### Correct Pattern
```rust
// Function only needs to read the name, not own the entire User
fn greet(name: &str) -> String {
    format!("Hello, {}", name)
}

// Accept slice instead of Vec - more flexible
fn calculate_average(values: &[f64]) -> f64 {
    values.iter().sum::<f64>() / values.len() as f64
}

// Use trait bound for flexibility
fn print_items<T: Display>(items: &[T]) {
    for item in items {
        println!("{}", item);
    }
}

// Called with narrow argument
greet(&user.name);
```

#### Incorrect Pattern
```rust
// Function receives entire User but only uses name
fn greet(user: &User) -> String {
    format!("Hello, {}", user.name)
}

// Requires Vec when slice would suffice
fn calculate_average(values: Vec<f64>) -> f64 {
    values.iter().sum::<f64>() / values.len() as f64
}

// Takes ownership unnecessarily
fn greet(name: String) -> String {
    format!("Hello, {}", name)
}
```

### Newtype Pattern for Primitive Arguments
When a function accepts two or more consecutive parameters of the same primitive type, use newtypes to distinguish them. This prevents argument transposition errors and makes the type signature self-documenting.

#### Principles
- Single primitive parameter is fine without newtype (e.g., `fn greet(name: &str)`)
- Two or more consecutive parameters of the same primitive type require newtypes
- Consider using a struct when there are more than 3 related primitive arguments
- Newtypes also enable implementing traits and adding validation

#### Correct Pattern
```rust
// Newtypes for domain concepts
struct Email(String);
struct Password(String);

// Two consecutive Strings - newtypes required
fn login(email: &Email, password: &Password) -> Result<AuthToken, AuthError> {
    ...
}

struct UserId(u64);
struct OrderId(u64);

// Two consecutive u64s - newtypes required
fn get_order(user_id: UserId, order_id: OrderId) -> Result<Order, Error> {
    ...
}

// Single primitive - newtype not required
fn greet(name: &str) -> String {
    format!("Hello, {}", name)
}

// Alternative: use struct for multiple related arguments
struct UserFilters {
    is_admin: bool,
    is_active: bool,
    min_age: u32,
}

fn filter_users(filters: &UserFilters, users: &[User]) -> Vec<&User> {
    ...
}

// Newtype with validation
struct PortNumber(u16);

impl PortNumber {
    fn new(port: u16) -> Result<Self, ValidationError> {
        if port != 0 {
            Ok(Self(port))
        } else {
            Err(ValidationError::InvalidPort)
        }
    }
}
```

#### Incorrect Pattern
```rust
// Two consecutive Strings - easy to transpose arguments
fn login(email: &str, password: &str) -> Result<AuthToken, AuthError> {
    ...
}

// Calling code is ambiguous
login("password123", "user@email.com")?;  // Oops! Arguments swapped

// Two consecutive u64s - impossible to distinguish
fn get_order(user_id: u64, order_id: u64) -> Result<Order, Error> {
    ...
}

// At call site, meaning is unclear
get_order(12345, 67890)?;  // Which ID is which?

// Multiple bools are confusing
fn filter_users(is_admin: bool, is_active: bool, include_deleted: bool) -> Vec<User> {
    ...
}

filter_users(true, false, true);  // What does this mean?
```

#### When to Apply
- Two or more consecutive `String`/`&str` parameters
- Two or more consecutive numeric parameters (`i32`, `u64`, `f64`, etc.)
- Two or more consecutive `bool` parameters
- Any ID types that could be confused (`user_id`, `order_id`, etc.)

### Make Invalid States Unrepresentable
Design types so that invalid combinations of data cannot be constructed. The compiler should prevent invalid states rather than relying on runtime checks.

#### Principles
- Use enums to model exact valid states
- Avoid Boolean blindness (multiple bool fields that have dependent meanings)
- Replace Option with explicit states when absence has semantic meaning
- Use NonZero types, NonEmpty collections when constraints exist
- Leverage the type system for compile-time guarantees

#### Correct Pattern
```rust
// Explicit states for async operations
enum RemoteData<T, E> {
    NotAsked,
    Loading,
    Success(T),
    Failure(E),
}

// User is either authenticated or anonymous - no invalid combinations
enum User {
    Authenticated { id: u64, name: String, email: String },
    Anonymous,
}

// Connection states are explicit
enum ConnectionState {
    Disconnected,
    Connecting { attempt: u32 },
    Connected { socket: TcpStream },
    Reconnecting { attempt: u32, last_error: Error },
}

// Non-empty guaranteed by type
struct NonEmptyVec<T> {
    first: T,
    rest: Vec<T>,
}

// Use NonZero for values that must not be zero
use std::num::NonZeroU32;

struct Pagination {
    page: NonZeroU32,
    per_page: NonZeroU32,
}

// Builder pattern for complex construction
struct ServerConfig {
    host: String,
    port: u16,
    // ... many fields
}

struct ServerConfigBuilder {
    // Optional fields during building
}

impl ServerConfigBuilder {
    fn build(self) -> Result<ServerConfig, ConfigError> {
        // Validate all fields before constructing
    }
}
```

#### Incorrect Pattern
```rust
// Multiple Options create invalid states
struct ApiResponse {
    is_loading: bool,
    data: Option<Data>,
    error: Option<String>,
}
// What if is_loading = true AND data = Some(x)?

// Boolean blindness
struct User {
    name: String,
    is_logged_in: bool,
    is_admin: bool,
    is_guest: bool,  // What does is_logged_in = true, is_guest = true mean?
}

// Impossible to know if empty is valid or error
struct SearchResults {
    items: Vec<Item>,
    has_searched: bool,
}

// Stringly typed - anything goes
struct Config {
    mode: String,  // Should be enum: Development, Production, Test
    log_level: String,  // Should be enum: Debug, Info, Warn, Error
}
```

#### Review Questions
- Can this type represent a state that should never occur?
- Are there multiple fields that must change together to remain consistent?
- Would an enum make the valid states more explicit?
- Are bool fields hiding meaningful domain states?
- Can invalid values be constructed at all?

### Lifetime and Borrowing Complexity
Complex lifetime annotations can indicate overly coupled code:

| Metric | Limit | Notes |
|--------|-------|-------|
| Lifetime parameters | ≤ 2 | More indicates complex ownership |
| Nested references | ≤ 2 levels | `&&T` is usually fine, `&&&T` needs review |

#### When Complexity is Acceptable
- Zero-copy parsing libraries
- Self-referential structures (with proper abstractions like `Pin`)
- Performance-critical code paths (documented)

### Unsafe Code Guidelines

| Metric | Limit | Notes |
|--------|-------|-------|
| Unsafe blocks per module | Minimize | Each must have SAFETY comment |
| Lines in unsafe block | ≤ 10 | Keep unsafe surface minimal |
| Unsafe functions | Document invariants | All preconditions must be documented |

#### Correct Pattern
```rust
/// # Safety
/// - `ptr` must be valid and properly aligned
/// - `ptr` must point to initialized memory
/// - The memory must not be accessed through any other pointer during this call
unsafe fn dangerous_operation(ptr: *mut u8) {
    // SAFETY: Caller guarantees ptr is valid and exclusive
    *ptr = 42;
}
```

### Code Review Checklist

#### Complexity
- [ ] Helper functions: cyclomatic complexity ≤ 12
- [ ] Helper functions: length < 50 lines
- [ ] Helper functions: match arms ≤ 10
- [ ] All functions: nesting depth ≤ 4
- [ ] All functions: parameters ≤ 5 (or use struct)
- [ ] Match dispatch arms: ≤ 5 lines or delegated to helper
- [ ] Iterator chains: ≤ 7 method calls
- [ ] Module length: < 1000 lines
- [ ] Lifetime parameters: ≤ 2 per function

#### Type Design
- [ ] Functions accept narrowest possible arguments
- [ ] References used instead of owned values where appropriate
- [ ] Slices used instead of Vec where appropriate
- [ ] Consecutive primitive parameters of the same type use newtypes
- [ ] Invalid states are unrepresentable
- [ ] No Boolean blindness (related bool fields)
- [ ] Enums used for explicit state modeling
- [ ] Option only used when absence is truly optional, not a distinct state

#### Safety
- [ ] Unsafe blocks minimized and documented with SAFETY comments
- [ ] All unsafe function preconditions documented
- [ ] No undefined behavior possible through safe API

### Exceptions
Document exceptions in code comments when limits are intentionally exceeded:
- FFI bindings requiring specific signatures
- Performance-critical code with benchmarks
- Generated code (macros, build scripts)
- Migration/legacy code with documented refactoring plan
