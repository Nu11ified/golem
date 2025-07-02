# golem-test-app

A Golem application - reactive web apps built with pure Go and WebAssembly.

## Getting Started

### Development
```bash
golem dev
```

### Build for Production
```bash
golem build
```

### Start Production Server
```bash
golem start
```

## Project Structure

- `src/app/` - Main application code
- `src/components/` - Reusable components  
- `src/server/` - Server functions (accessible via gRPC)
- `.golem/` - Build artifacts and type definitions

## Features

✅ **Pure Go** - No JavaScript required
✅ **Type Safety** - Full type safety from backend to frontend
✅ **Hot Reload** - Fast development with instant updates
✅ **Server Functions** - Call Go functions from the client via gRPC
✅ **Component System** - Reusable, composable UI components
✅ **Virtual DOM** - Efficient rendering with minimal updates
✅ **WebAssembly** - Near-native performance in the browser

## Learn More

- [Golem Documentation](https://github.com/Nu11ified/golem)
- [Go WebAssembly](https://github.com/golang/go/wiki/WebAssembly)
