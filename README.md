# 🖥 amsh (Ape Machine Shell)

## Technical Architecture

### Memory Systems

#### Neo4j Integration

```go
// Graph-based memory system
package tools

type Neo4j struct {
    client    neo4j.DriverWithContext
    Operation string // query/write
    Cypher    string
}
```

-   Persistent graph memory storage
-   Schema-driven query interface
-   Supports both read (query) and write operations
-   JSON-based response format

#### Qdrant Integration

-   Vector-based similarity search
-   Semantic memory storage
-   Embedding management
-   Nearest neighbor search capabilities

### Process Architecture

#### Abstract Processes

##### Fractal Process

```go
package fractal

type Process struct {
    BasePattern    Pattern  // Fundamental repeating pattern
    Scales         []Scale  // Multi-level manifestation
    Iterations     int      // Recursion depth
    SelfSimilarity float64  // Pattern preservation metric
}
```

-   Pattern-based reasoning
-   Scale-invariant processing
-   Self-similar structure recognition
-   Hierarchical pattern matching

##### Tensor Process

```go
package tensor

type Process struct {
    Dimensions   []Dimension   // Relationship space aspects
    TensorFields []TensorField // Multi-dimensional patterns
    Projections  []Projection  // Dimensional reductions
}
```

-   Multi-dimensional relationship modeling
-   Field-based pattern recognition
-   Dimensional projection capabilities
-   Dynamic tensor manipulation

### Agent Architecture

#### Main Agent

```go
package marvin

type Agent struct {
    ID        string
    ctx       context.Context
    buffer    *Buffer
    processes map[string]Process
    prompt    *Prompt
}
```

#### Sidekick Agents (Tool Handlers)

-   Dedicated process execution
-   Tool-specific operations
-   Focused context management
-   Independent buffer spaces

### Process Layering System

1. **Base Layer (Abstract Processes)**

    - Fractal pattern recognition
    - Tensor relationship modeling
    - Pattern extraction and matching

2. **Tool Layer**

    ```go
    // Tool interface implementation
    func (tool *Tool) Use(ctx context.Context, args map[string]any) string
    func (tool *Tool) GenerateSchema() string
    ```

    - Neo4j graph operations
    - Qdrant vector operations
    - Schema-driven tool execution

3. **Integration Layer**
    - Process output transformation
    - Memory system synchronization
    - Context propagation
    - Event stream management

### Context Management

#### Buffer System

```go
package marvin

type Buffer struct {
    messages         []provider.Message
    maxContextTokens int  // 128k token window
}
```

-   Token-aware truncation
-   Priority message preservation
-   Smart context window management
-   Message role structuring

#### Process Context Flow

1. Abstract process execution
2. Memory system integration
3. Tool-based augmentation
4. Context window management

### Event System

#### Generation Pipeline

```go
package marvin

func (agent *Agent) Generate() <-chan provider.Event {
    // Event stream setup
    // Process execution
    // Context management
    // Response accumulation
}
```

#### Provider Integration

-   Balanced provider selection
-   Event stream accumulation
-   Asynchronous processing
-   Response aggregation

### Schema-Driven Integration

#### Tool Schema Generation

```go
func GenerateSchema[T any]() string {
    // JSON schema reflection
    // Tool interface definition
    // Operation specification
}
```

#### Process Schema Layering

1. Base process schema definition
2. Tool operation schema
3. Response format schema
4. Integration validation schema

### Implementation Notes

1. **Process Layering Strategy**

    - Abstract processes feed into tool processes
    - Memory systems provide context enrichment
    - Sidekick agents handle specialized operations
    - Main agent orchestrates overall flow

2. **Memory Integration Pattern**

    - Graph-based relationship storage (Neo4j)
    - Vector-based similarity matching (Qdrant)
    - Hybrid memory access patterns
    - Context-aware memory operations

3. **Agent Collaboration Model**
    - Main agent delegates to sidekicks
    - Tool-specific context isolation
    - Shared memory access patterns
    - Event-based communication

### Agent-Provider Interaction Model

The system implements a sophisticated provider abstraction layer that manages interactions between agents and various AI providers. This architecture is designed for reliability, scalability, and fault tolerance.

#### Provider Interface

-   Common interface abstracting all AI providers (OpenAI, Anthropic, Google, Cohere)
-   Event-driven streaming communication pattern
-   Structured message format with role-based content modeling
-   Support for multiple event types:
    -   Token events (streaming responses)
    -   Tool call events (function execution)
    -   Error events (failure handling)
    -   Done events (completion signals)

#### Balanced Provider System

The system implements an intelligent load balancer for AI provider management:

```go
type ProviderStatus struct {
    name     string
    provider Provider
    occupied bool
    lastUsed time.Time
    failures int
    mu       sync.Mutex
}
```

Key Features:

-   Provider status tracking and health monitoring
-   Failure detection with cooldown periods
-   Occupation tracking to prevent overload
-   Intelligent provider selection based on:
    -   Current availability
    -   Historical performance
    -   Failure count
    -   Last usage time

#### Event-Driven Architecture

-   Asynchronous communication using Go channels
-   Structured event types for different response categories
-   Accumulator pattern for response aggregation
-   Non-blocking provider selection

### Container-Based Isolation

The system uses a robust container-based isolation model to ensure secure and reproducible execution environments.

#### Builder System

```go
type Builder struct {
    client *client.Client
}
```

Features:

-   Docker client abstraction
-   Context-aware build process
-   Output streaming and error handling
-   Build option management

#### Environment Management

The environment system provides:

-   Container lifecycle management
-   Command execution isolation
-   Environment variable handling
-   Workspace management

#### Security Model

1. Container Configuration:

    - Base: `bitnami/minideb` (minimal attack surface)
    - Dynamic user creation
    - Privilege separation
    - Workspace isolation

2. Security Features:

    ```shell
    # Dynamic user creation with controlled privileges
    USERNAME=${USERNAME:-devuser}
    useradd -m -s /bin/bash "$USERNAME"
    echo "$USERNAME ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/$USERNAME
    ```

    - Non-root execution
    - Controlled sudo access
    - SSH security configuration
    - Isolated workspace at `/tmp/workspace/amsh`

### Integration Points

The system integrates these components through several key mechanisms:

1. Provider Integration:

    - Container-isolated provider calls
    - Load-balanced provider selection
    - Failure recovery and retry logic

2. Security Boundaries:

    - Container isolation for execution
    - Provider API key management
    - Workspace separation

3. Communication Flow:
    ```
    Agent -> Balanced Provider -> Container -> AI Provider
       ^            |               |            |
       |            v               v            v
    Response <- Event Stream <- Execution <- API Call
    ```

### Development Considerations

1. Provider Management:

    - Monitor provider health and performance
    - Implement circuit breakers for failing providers
    - Consider adding provider-specific retry strategies

2. Container Security:

    - Regular security audits of base images
    - Implementation of resource limits
    - Proper secret management

3. Scalability:
    - Provider pool management
    - Container orchestration
    - Resource allocation strategies
