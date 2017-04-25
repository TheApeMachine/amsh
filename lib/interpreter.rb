class Interpreter

  def run(response, command)
    if response == 'escalate-command'
      send(command)
    else
      return "#{response}\n"
    end
  end

  def method_missing(m, *args, &block)
    begin
      res = `#{m.to_s}`
    rescue Errno::ENOENT
      begin
        require "./bin/#{m}"
        @m = Object.const_get(m.to_s.capitalize).new
        @m.run
      rescue LoadError
        return "Command failed!\n"
      end
    end
  end

end
