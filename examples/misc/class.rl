class Car 
  def initialize(type) {
    @type = type
    @_distance = 0
    
    //puts("initialized car)
  }

  def _is_broken() {
    return (@distance > 5)
  }

  //def drive()
  //  @distance = @distance + 1
  //end

  //def self.types()
  //  return ["Volvo", "BMW", "Porsche"]
  //end
end

test = Car.new()