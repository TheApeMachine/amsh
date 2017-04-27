class Interpreter

  def initialize
    @ai = Ai.new
  end

  def run(response, command)
    send(command)
  end

  def method_missing(m, *args, &block)
    begin
      res = `#{m.to_s}`
    rescue Errno::ENOENT
      begin
        require "./bin/#{m}"
        @m = Object.const_get(m.to_s.capitalize).new(args)
        @m.run
      rescue LoadError
        @ai.run("#{m} #{args.join(' ')}")
      end
    end
  end

end
