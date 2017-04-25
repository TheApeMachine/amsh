class Interpreter

  def run(response, command)
    if response == 'escalate-command'
      send(command)
    else
      return response
    end
  end

  def method_missing(m, *args, &block)
    require "./bin/#{m}"
    @m = Object.const_get(m.to_s.capitalize).new
    @m.run
  end

end
