# CLAUDE.md — Project Guidelines

This file documents conventions and guidelines for working on this project with Claude.

---

## Git Conventions

Branch naming, commit message format, PR standards, and merge strategy based on
[telvie-codex guidelines](https://github.azc.ext.hp.com/telvie/telvie-codex/blob/main/docs/general/git/).

### Branch Naming Convention

Format: `<type>/<short>_<feature>_(<ticket_id>)`

**Rules:**
- DO NOT use user name as branch name
- Use lowercase letters
- Use underscores `_` to separate words
- Keep names short but meaningful
- Optionally include ticket or issue ID

**Types:** `feat`, `fix`, `refactor`, `docs`, `test`, `build`, `ci`, `perf`, `hotfix`, `release`

**Examples:**
- `feat/email_verification_8520`
- `fix/empty_profile_image_6636`
- `refactor/data_fetching_logic`
- `docs/contributing_guide`
- `test/signup_edge_cases`
- `release/25Q2.5`
- `hotfix/api_key_leak`

### Commit Message Format

Format: `<type>(<scope>): <short summary> (<ticket number>)`

**Mandatory fields:** `<type>` and `<summary>`  
**Optional fields:** `(<scope>)` and `(<ticket number>)`

#### Type

- `build`: Changes that affect the build system or external dependencies
- `ci`: Changes to CI configuration files and scripts
- `docs`: Documentation only changes
- `feat`: A new feature
- `fix`: A bug fix
- `perf`: A code change that improves performance
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `test`: Adding missing tests or correcting existing tests
- `revert`: Commit reverts a previous commit

#### Summary Rules

- Use imperative, present tense: "change" not "changed" nor "changes"
- Don't capitalize the first letter
- No dot (.) at the end
- Keep it short (ideally under 70 characters)
- Include ticket number from ADO (AB#1234) or Jira (xxxx-1234) if applicable

**Examples:**
- `feat(user): add user registration flow AB#8520`
- `fix(auth): fix missing token on login AB#1234`
- `docs: update getting started section`
- `refactor(api): simplify request/response logic`
- `perf(home): reduce render blocking resources`
- `test(user): add login unit tests`
- `build(deps): upgrade axios to v1.6.5`
- `ci(github): add deploy to production workflow`

#### Revert Commits

If reverting a previous commit, format as: `revert: <original commit header>`

The commit message body should contain:
- `This reverts commit <SHA>`
- A clear description of the reason for reverting

### Pull Request

The PR title must follow the same format as Git commit messages.

### Merge Strategy

**Recommended: Squash and Merge**

- Squashing combines all commits into a single commit, keeping history clean and concise
- Makes it easier to understand the changes made in a feature or bug fix
- Reduces clutter in the commit history

### Key Rules

- **NEVER commit directly to main branch**
- Create a new branch from `main` for every feature or bug fix
- At least one other developer must review the PR before merging
- Use interactive rebase (`git rebase -i`) if you need to clean up commit history before merging
