Cache layer generator for Go

1. Structs (models) -> .proto files +
2. Converts (models) -> proto & vice versa - create converters
3. Generate interfaces for Queries +
4. Generate implementation for Queries that accepts cache 
5. Configuration:
5.1. sources (models, Queries?)
5.2. overrides for types (converters)
5.3. specify: cache key, cache ttl, cache prefix (get/or update/mget?), ignore specific methods
    

entity map (shorten keys)