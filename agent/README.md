# Design

## Package
* This package puts some concrete configurations

## Modules
    |+algorithm
    |+cs
    |+cnf
    |+monitoring
    |+transshipment    // handle the linker
    |+redis            // handle the backend: Pool & Redis 

## Guide of handling error

* Logging errors in the bottom layer
* Let logical layer handle errors
* If the error in the bottom layer can be handled, do it right now
