# Browse Together

## Development
- Redis : `docker run -d --name redis-stack-server -p 6379:6379 redis/redis-stack-server:latest`
- Setup submodules: `git submodule update --init --recursive`
- Start frontend example page: `cd browse-together-example-page && npm run dev`
- Start Backend Server : `ENV=dev go run server.go`


## linking sdk to example page
- SDK: npm run build
- EXAMPLE WEB: npm install ../browse-together-sdk

```
   <script type="module">
      import {BrowseTogetherSdk} from '/static/js/index.mjs'
      const bts = new BrowseTogetherSdk("http://localhost:8080", [
        { classList: [], id: undefined, tag: 'section' },
      ]);
      console.log(bts.ping())
    </script>
```