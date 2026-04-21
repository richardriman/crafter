## Elm Code Complexity Rules

### Overview
Elm's design enforces simplicity through its pure functional nature, strong type system, and exhaustive pattern matching. However, these features can artificially inflate traditional complexity metrics. The following rules account for Elm-specific patterns.

### Standard Functions
All helper functions, utility functions, and non-routing functions must adhere to these limits:

| Metric | Limit | Notes |
|--------|-------|-------|
| Cyclomatic complexity | ≤ 15 | Pattern matching and Maybe/Result increase branch count |
| Cognitive complexity | ≤ 12 | Nested case/let expressions reduce readability |
| Function length | < 30 lines | Excluding type signatures and comments |
| Case branches | ≤ 8 | Per single case expression |
| Nesting depth | ≤ 3 levels | Nested case, let, if blocks |
| Parameters | ≤ 4 | Use Record type for more parameters |
| Pipeline length | ≤ 7 operators | Split longer |>/`<|` chains into named steps |

### TEA Routing Functions (update, view dispatchers, subscriptions)
These functions act as routers and are exempt from branch and length limits:

| Metric | Limit | Notes |
|--------|-------|-------|
| Case branches | No limit | Function acts as message router |
| Function length | No limit | Delegates logic to helpers |
| **Single branch length** | ≤ 5 lines | Inline code per branch |
| **Branch complexity** | ≤ 1 expression | Extract to helper function if more complex |

#### Correct Pattern
```elm
update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        ClickedLogin credentials ->
            handleLogin credentials model

        ClickedLogout ->
            handleLogout model

        GotApiResponse result ->
            handleApiResponse result model

        FormFieldChanged field value ->
            handleFormChange field value model
```

#### Incorrect Pattern
```elm
update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        ClickedLogin credentials ->
            let
                validatedCredentials =
                    validateCredentials credentials

                newModel =
                    { model | isLoading = True, errors = [] }
            in
            if isValid validatedCredentials then
                ( newModel, Api.login validatedCredentials )
            else
                ( { model | errors = getErrors validatedCredentials }, Cmd.none )
        -- ... more branches with inline logic
```

### Module-Level Limits

| Metric | Limit | Action when exceeded |
|--------|-------|----------------------|
| Module length | < 1000 lines | Split into submodules |
| Msg variants per module | ≤ 25 | Consider Page/Feature submodules |
| Exposed functions | ≤ 15 | Module may have too many responsibilities |

### Narrowing Function Arguments
Functions should accept the narrowest possible type that satisfies their requirements. This reduces coupling, improves testability, and makes the function's actual dependencies explicit.

#### Principles
- Pass only the data the function actually needs, not the entire Model
- Use specific types instead of generic ones when possible

#### Correct Pattern
```elm
-- Function only needs user's name, not entire User record
viewGreeting : String -> Html msg
viewGreeting userName =
    h1 [] [ text ("Hello, " ++ userName) ]

-- Called from update/view with narrow argument
viewGreeting model.user.name
```

#### Incorrect Pattern
```elm
-- Function receives entire Model but only uses one field
viewGreeting : Model -> Html msg
viewGreeting model =
    h1 [] [ text ("Hello, " ++ model.user.name) ]

-- Function receives entire User but only needs name
viewGreeting : User -> Html msg
viewGreeting user =
    h1 [] [ text ("Hello, " ++ user.name) ]
```

#### Exceptions
- TEA routing functions (update, view) naturally receive the full Model
- Functions that genuinely need many fields from a record

### Semantic Type Aliases for Consecutive Primitive Arguments
When a function accepts two or more consecutive parameters of the same primitive type (String, Int, Bool), use type aliases to distinguish them. This prevents argument transposition errors and makes the type signature self-documenting - especially important when functions are passed as arguments without visible parameter names.

#### Principles
- Single primitive parameter is fine without type alias (e.g., `String -> Html msg`)
- Two or more consecutive parameters of the same primitive type require type aliases
- Consider using a Record type when there are more than 3 related primitive arguments

#### Correct Pattern
```elm
type alias Email =
    String

type alias Password =
    String

-- Two consecutive Strings - type aliases required
login : Email -> Password -> Cmd Msg
login email password =
    ...

type alias IsAdmin =
    Bool

type alias IsActive =
    Bool

-- Two consecutive Bools - type aliases required
filterUsers : IsAdmin -> IsActive -> List User -> List User
filterUsers isAdmin isActive users =
    ...

-- Single String - type alias not required
viewGreeting : String -> Html msg
viewGreeting userName =
    h1 [] [ text ("Hello, " ++ userName) ]

-- Alternative: use Record for multiple related arguments
type alias UserFilters =
    { isAdmin : Bool
    , isActive : Bool
    , minAge : Int
    }

filterUsers : UserFilters -> List User -> List User
filterUsers filters users =
    ...
```

#### Incorrect Pattern
```elm
-- Two consecutive Strings - easy to transpose arguments
login : String -> String -> Cmd Msg
login email password =
    ...

-- Calling code is ambiguous
login "password123" "user@email.com"  -- Oops! Arguments swapped

-- Two consecutive Bools - impossible to understand from type signature alone
filterUsers : Bool -> Bool -> List User -> List User
filterUsers isAdmin isActive users =
    ...

-- At call site, meaning is unclear
filterUsers True False users  -- Which Bool is which?

-- When passed as argument, parameter names are not visible
login : String -> String -> Cmd Msg
login =
    Auth.login -- Only type signature is visible, not parameter names
```

#### When to Apply
- Two or more consecutive String parameters
- Two or more consecutive Int parameters
- Two or more consecutive Bool parameters
- Two or more consecutive Float parameters
- Two or more consecutive numbers
- Two or more consecutive dates
- Two or more consecutive same types in general

### Make Invalid States Unrepresentable
Design types so that invalid combinations of data cannot be constructed. The compiler should prevent invalid states rather than relying on runtime checks.

#### Principles
- Use custom types to model exact valid states
- Avoid Boolean blindness (multiple Bool fields that have dependent meanings)
- Replace Maybe with explicit states when absence has semantic meaning
- Use non-empty types (e.g., `( first, List rest )`) when lists must have at least one element

#### Correct Pattern
```elm
-- Explicit states for remote data
type RemoteData e a
    = NotAsked
    | Loading
    | Failure e
    | Success a

-- User is either logged in with data or logged out - no invalid combinations
type User
    = LoggedIn { id : Int, name : String, email : String }
    | Guest

-- Form states are explicit
type FormState
    = Editing FormData
    | Submitting FormData
    | SubmitFailed FormData (List String)
    | SubmitSucceeded

-- Non-empty list guaranteed by type
type alias NonEmptyList a =
    ( a, List a )

selectedItems : NonEmptyList Item -> Html msg
```

#### Incorrect Pattern
```elm
-- Multiple Maybes create invalid states (what if isLoading = True and data = Just x?)
type alias Model =
    { isLoading : Bool
    , data : Maybe Data
    , error : Maybe String
    }

-- Boolean blindness - what does isAdmin = True, isGuest = True mean?
type alias User =
    { name : String
    , isLoggedIn : Bool
    , isAdmin : Bool
    , isGuest : Bool
    }

-- Impossible to know if empty list is valid or an error
type alias Model =
    { items : List Item
    , hasLoadedItems : Bool
    }
```

#### Review Questions
- Can this type represent a state that should never occur?
- Are there multiple fields that must change together to remain consistent?
- Would a custom type make the valid states more explicit?
- Are Boolean fields hiding meaningful domain states?

### Code Review Checklist

#### Complexity
- [ ] Helper functions: cyclomatic complexity ≤ 15
- [ ] Helper functions: length < 30 lines
- [ ] Helper functions: case branches ≤ 8
- [ ] All functions: nesting depth ≤ 3
- [ ] All functions: parameters ≤ 4 (or use Record)
- [ ] Update/view branches: ≤ 5 lines or delegated to helper
- [ ] Pipeline chains: ≤ 7 operators
- [ ] Module length: < 1000 lines
- [ ] Msg variants: ≤ 25 per module (document exceptions)

#### Type Design
- [ ] Functions accept narrowest possible arguments
- [ ] No unnecessary Model/Record passing
- [ ] Consecutive primitive parameters of the same type have semantic type aliases
- [ ] Invalid states are unrepresentable
- [ ] No Boolean blindness (related Bool fields)
- [ ] Custom types used for explicit state modeling
- [ ] Maybe only used when absence is truly optional, not a distinct state

### Exceptions
Document exceptions in code comments when limits are intentionally exceeded:
- Large Msg types representing complex domain logic
- Generated code or decoders
- Migration/legacy code with documented refactoring plan
