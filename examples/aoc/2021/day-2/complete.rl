input = open("examples/aoc/2021/day-2/input").lines()

depth = 0
hor = 0
aim = 0

foreach i, line in input {
  command = line.split(" ")[0]
  value = line.strip().split(" ")[1].plz_i()
  if (command == "forward") {
    hor = hor + value
    depth = depth + (value * aim)
  }

  if (command == "down") {
    aim = aim + value
  }

 if (command == "up") {
    aim = aim - value
  }
}

puts(hor * depth)