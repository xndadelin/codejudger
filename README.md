# codejudger

> high-performance code execution engine for hackacode (omggg)

## what's this?

codejudger is a robust, sandboxed code execution system designed to power the hackacode competitive programming platform. built with go and utilizing the isolate sandbox, it provides secure code execution, test case validation, and performance metrics for multiple programming languages.

## features

- secure sandboxed execution - run untrusted code safely with isolate
- multi-language support - judge solutions in c++, python, javascript, ruby, php, and c#
- performance metrics - accurate measurement of execution time and memory usage
- test case validation - automatically verify code against test cases
- rest api - simple integration with web applications
- swagger documentation - easy-to-use api reference
- supabase integration - seamless database connectivity

## getting started

1. install isolate sandbox (requires linux):
   ```bash
   git clone https://github.com/ioi/isolate.git
   cd isolate
   make
   sudo make install
   ```

2. configure environment variables:
   ```bash
   cp .env.example .env
   # edit the .env file with required settings:
   # - JWT_SECRET: secret key for JWT token generation
   # - URL: your Supabase project URL
   # - ANON_API_KEY: your Supabase anonymous key
   # - ENVIRONMENT: development or production
   ```

3. build and run the server:
   ```bash
   go build -o codejudger ./cmd/server
   chmod +x ./codejudger
   ./codejudger
   ```

4. access the swagger documentation:
   ```
   http://localhost:8080/swagger/index.html
   ```

## api usage

```bash
# run code with custom input
curl -X POST http://localhost:1072/api/v1/run \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-jwt-token" \
  -d '{
    "code": "print(input())",
    "language": "Python", 
    "input": "Hello World!"
  }'

# submit a solution for judging
curl -X POST http://localhost:1072/api/v1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-jwt-token" \
  -d '{
    "code": "...",
    "language": "C++",
    "slug": "two-sum"
  }'

# get a JWT token from your API key
curl -X POST http://localhost:1072/get-token \
  -H "Content-Type: application/json" \
  -d '{
    "api_key": "your-api-key"
  }'
```

## tech stack

- go - high-performance backend language
- isolate - secure sandboxing system (used in competitive programming contests like IOI)
- swagger - api documentation
- supabase - database and authentication
- docker - containerization for easy deployment
- coolify - self-hosted deployment platform for simple deployments

## development

```bash
# run in development mode
go run ./cmd/server/main.go

# build for production
go build -o codejudger ./cmd/server

# update swagger documentation
swag init -g cmd/server/main.go
```

## contributing
contributions are welcome! please feel free to submit a pull request. don't worry about strict formatting or tests - just code however you want lol, as long as it works. the only real rule is to keep the vibes good!

## license

codejudger is open-source software licensed under the mit license.

## credits

built during neighborhood by Adelin!!
