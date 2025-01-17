---
title: "Null"
menu:
  docs:
    parent: "literals"
---
# Null




## Literal Specific Methods

### plz_f()
> Returns `FLOAT`

Returns zero float.



### plz_i()
> Returns `INTEGER`

Returns zero integer.



### plz_s()
> Returns `STRING`

Returns empty string.




## Generic Literal Methods

### methods()
> Returns `ARRAY`

Returns an array of all supported methods names.

```js
🚀 > "test".methods()
=> [count, downcase, find, reverse!, split, lines, upcase!, strip!, downcase!, size, plz_i, replace, reverse, strip, upcase]
```

### type()
> Returns `STRING`

Returns the type of the object.

```js
🚀 > "test".type()
=> "STRING"
```

### wat()
> Returns `STRING`

Returns the supported methods with usage information.

```js
🚀 > true.wat()
=> BOOLEAN supports the following methods:
				plz_s()
```
