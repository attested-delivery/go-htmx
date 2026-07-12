## Summary

<!-- Brief description of the changes in this PR -->

## Related Issues

<!-- Link any related issues: Fixes #123, Relates to #456 -->

## Changes

<!-- List the key changes made in this PR -->

-
-
-

## Type of Change

<!-- Mark the relevant option with an [x] -->

- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to change)
- [ ] Documentation update
- [ ] Refactoring (no functional changes)
- [ ] Performance improvement
- [ ] Test coverage improvement

## Testing

<!-- Describe how you tested these changes -->

### Test Commands Run

```bash
go build ./...
go vet ./...
gofmt -l .
go test ./... -race -cover
govulncheck ./...
```

### Manual Testing

<!-- Describe any manual testing performed -->

## Checklist

<!-- Mark completed items with an [x] -->

### Code Quality

- [ ] My code follows the project's code style (`gofmt`)
- [ ] I have run `go vet` and resolved all issues
- [ ] No unhandled errors (no bare `_ = err` without justification)

### Testing

- [ ] I have added tests that prove my fix/feature works
- [ ] New and existing tests pass locally with `go test ./...`

### Documentation

- [ ] I have updated the documentation accordingly
- [ ] I have added doc comments for new exported identifiers
- [ ] I have updated the CHANGELOG.md (if applicable)

### Supply Chain

- [ ] I have run `govulncheck ./...` and resolved any issues
- [ ] New dependencies are justified and from trusted sources
- [ ] No new security advisories are introduced

### Commit Hygiene

- [ ] My commits follow conventional commit format
- [ ] I have rebased on the latest main branch
- [ ] I have squashed fixup commits

## Performance Impact

<!-- If applicable, describe any performance implications -->

## Screenshots

<!-- If applicable, add screenshots to help explain your changes -->

## Additional Notes

<!-- Add any additional context about the PR here -->
