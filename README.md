# git-monitor

Git Monitor is a tool for tracking GitHub repositories and monitoring commits. It allows you to add repositories to track, fetch commit data, and analyze commit statistics through a set of RESTful API endpoints.

---

## Intro

Git Monitor provides a backend API for monitoring Git repositories. It enables you to:

- **Add repositories for tracking:** Easily add a GitHub repository to monitor for new commits.
- **List and update repository settings:** Retrieve your tracked repositories and adjust settings.
- **Trigger background tasks:** Manually or automatically initiate tasks to fetch and process commit data.
- **Retrieve commit details and statistics:** Get information about commits and top commit authors.

The API endpoints are designed to be intuitive and RESTful, with clear error messages and HTTP status codes to aid in development and troubleshooting.

---

## Setup

### Prerequisites

- [Go](https://golang.org/) installed on your machine.
- [Docker](https://www.docker.com/) Docker installed and running on your machine.
- Required dependencies installed as per the project documentation.

### Installation & Running

1. **Clone the Repository:**

```
   git clone https://github.com/yourusername/git-monitor.git cd git-monitor
```

2. **Build and Run the Application:**

Use the provided Makefile commands:

- **Build and run:**
  ```
  make build-run
  ```
- **Or, run locally with:**
  ```
  make run
  ```

3. **Database Setup:**

- **Run migrations:**
  ```
  make migrate
  ```
- **Reset the database (deletes the current database file and reapplies migrations):**
  ```
  make reset-db
  ```

4. **Testing:**

Run tests with:

```
   make test
```


---

## Dependencies

### Database

- **MySQLite:** A lightweight file-based database used for storing repository and commit data.

### Packages

- **testify:** Testing framework for unit tests.
- **migrator-tool:** Tool used for managing and applying database migrations.

Other internal packages include:

- **githubclient:** For interacting with the GitHub API.
- **logger:** For structured logging.
- **smq:** A simple message queue for task event handling.

---

## Schema

### Repositories

| Field                     | Type   | Description                                              | Sample Value                                      |
| ------------------------- | ------ | -------------------------------------------------------- | ------------------------------------------------- |
| `ID`                      | string | Unique identifier for the repository                     | `repo-12345678`                                   |
| `RepoID`                  | string | GitHub repository ID                                     | `1234567`                                         |
| `Name`                    | string | Repository name                                          | `git-monitor`                                     |
| `Owner`                   | string | Repository owner's login                                 | `victor-nach`                                     |
| `Description`             | string | Description of the repository                            | "Monitors Git commits"                            |
| `URL`                     | string | URL of the GitHub repository                             | `https://github.com/victor-nach/git-monitor`        |
| `Language`                | string | Primary programming language                             | `Go`                                              |
| `ForksCount`              | int    | Number of forks                                          | `5`                                               |
| `StarsCount`              | int    | Number of stars                                          | `10`                                              |
| `OpenIssues`              | int    | Number of open issues                                    | `2`                                               |
| `WatchersCount`           | int    | Number of watchers                                       | `8`                                               |
| `IsSyncedToStartTime`     | bool   | Indicates if syncing started from a set start time       | `false`                                         |
| `IsActive`                | bool   | Flag indicating if the repository is actively tracked    | `true`                                            |
| `CommitTrackingStartTime` | time   | Start time for commit tracking                           | `2021-01-01T00:00:00Z`                            |
| `LastFetchedAt`           | time   | Last time commits were fetched                           | `2021-02-01T00:00:00Z`                            |
| `RepoCreatedAt`           | time   | Date the repository was created                          | `2020-12-01T00:00:00Z`                            |
| `RepoUpdatedAt`           | time   | Date the repository was last updated                     | `2021-03-01T00:00:00Z`                            |
| `CreatedAt`               | time   | Timestamp when the repository was added to tracking      | `2021-03-15T00:00:00Z`                            |
| `UpdatedAt`               | time   | Timestamp when the repository record was last updated    | `null`                                          |

### Commits

| Field          | Type   | Description                                      | Sample Value                                          |
| -------------- | ------ | ------------------------------------------------ | ----------------------------------------------------- |
| `ID`           | string | Unique commit identifier                         | `commit-123456789`                                    |
| `SHA`          | string | Git commit SHA                                   | `a1b2c3d4`                                            |
| `RepositoryID` | string | Foreign key linking to the repository            | `repo-12345678`                                       |
| `RepoName`     | string | Name of the repository                           | `git-monitor`                                         |
| `Message`      | string | Commit message                                   | "Fix bug in API"                                      |
| `URL`          | string | URL to the commit on GitHub                      | `https://github.com/victor-nach/git-monitor/commit/a1b2c3d4` |
| `Author`       | string | Name of the commit author                        | `John Doe`                                            |
| `AuthorEmail`  | string | Email address of the commit author               | `john@example.com`                                    |
| `Date`         | time   | Date of the commit                               | `2021-03-14T12:00:00Z`                                  |
| `CreatedAt`    | time   | Timestamp when the commit was recorded           | `2021-03-14T12:05:00Z`                                  |
| `UpdatedAt`    | time   | Timestamp when the commit was last updated       | `null`                                              |

### Tasks

| Field      | Type   | Description                                      | Sample Value                |
| ---------- | ------ | ------------------------------------------------ | --------------------------- |
| `ID`       | string | Unique identifier for a scheduled task           | `task-123456789`            |
| `RepoNames`| array  | List of repository names associated with the task | `["git-monitor"]`           |
| `CreatedAt`| time   | Timestamp when the task was created               | `2021-03-14T12:10:00Z`       |
| `Status`   | string | Current status of the task (e.g., in-progress, completed) | `completed`             |

---

## Endpoints

### Repositories

1. **Add a new repository to track.**
- **POST `api/v1/repos/:owner/:repo?since=start-time`**  
____

2. **Retrieve a list of tracked repositories.**
- **GET  `/api/v1/repos/`**  
_______

3. Reset a repo, delete old records and start again
- **POST  `/repositories/:repo_name/reset`**  
_____

### Commits

- **GET  `/repos/:owner/:repo/top-authors`** 
Get the top commit authors for a repository.
__

- **GET  `/repos/:owner/:repo/commits`** 
Get the top commit authors for a repository.

___

## API Errors

Specific Error Messages:

- invalid request body
- failed to add tracked repository
- failed to list tracked repositories
- failed to update repository settings
- failed to get task
- failed to trigger task
- tracked repository not found
- duplicate repository

_________

## Postman collection
A Postman collection is provided in the root of the project as .postman_collection.json. You can use this collection to quickly explore and test the API endpoints.

To use the Postman collection:

Open Postman.
Click on "Import" and select the .postman_collection.json file from the project root.
Explore the endpoints included in the collection and run sample requests.