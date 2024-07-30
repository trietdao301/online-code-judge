# Coodbox

A web application similar to LeetCode where users can submit code, solve problems, and administrators can manage the platform.

## Features

- [x] User authentication using JWT
- [x] Role-based authorization (Contestant, Admin, Problem Setter)
- [x] Problem listing
  - [x] View all problems
  - [x] Filter problems
- [x] Problem management (Admin/Problem Setter)
  - [x] Create new problems
  - [x] Edit existing problems
  - [x] Create submission templates
- [x] Test case management (Admin/Problem Setter)
  - [x] Create test cases
  - [x] Edit test cases
- [x] Code submission
  - [x] Submit code for a problem
  - [x] View submission history
- [x] Submission judging
  - [x] Execute code in Docker containers
  - [x] Use Docker API for container management
- [x] Results and feedback
  - [x] View submission results
  - [x] Access execution logs
- [] Support languages
  - [x] Python
  - [x] Java

## TODO

- [ ] Add message queue for submission execution
- [ ] Implement user profile management
- [ ] Add support for more programming languages
- [ ] Create a leaderboard system
- [ ] Implement a discussion forum for each problem
- [ ] Create an admin dashboard for platform statistics
- [ ] Add support for custom input testing
- [ ] Implement a plagiarism detection system
- [ ] Add new theme

## Technologies Used

- Backend: [Golang]
- Frontend: [Typescript + React]
- Database: [MongoDB]
- Authentication: JWT
- Container: Docker
