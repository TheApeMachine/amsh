class Ai

  def initialize
    @client = ApiAiRuby::Client.new(
      client_access_token: '24de8cfec3d7493e9d7208e8e72f8fcd',
      api_session_id: rand(9999999)
    )
  end

  def run(command)
    response = @client.text_request(command)[:result][:fulfillment][:speech]
    send(response)
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
        return "#{m} #{args.join(' ')}"
      end
    end
  end

end
