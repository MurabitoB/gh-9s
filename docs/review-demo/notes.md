# Review demo

A file for trying out the review screen.

## Steps

1. Open the diff
2. Comment on a line
3. Send the review
4. Check it on the PR page

## Settings

| name  | default |
|-------|---------|
| loud  | false   |
| count | 0       |
| quiet | false   |

Nothing else to see here.

## Example

```go
g := &demo.Greeter{Name: "ada", Loud: true}
fmt.Println(g.Greet("world"))
```
