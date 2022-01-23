class Car 
  def initialize(type)
    private @distance
    @type = type
    @distance = 0
    
    puts("initialized car)
  end

  private def is_broken()
    return (@distance > 5)
  end

  def drive()
    @distance = @distance + 1
  end

  def self.types()
    return ["Volvo", "BMW", "Porsche"]
  end
end