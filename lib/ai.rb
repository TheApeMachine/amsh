class Ai

  def initialize
    @client = ApiAiRuby::Client.new(
      client_access_token: '24de8cfec3d7493e9d7208e8e72f8fcd',
      api_session_id: 0
    )
  end

  def run(command)
    return @client.text_request(command)[:result][:fulfillment][:speech]
  end

end
