class Ai

  def initialize
    @client = ApiAiRuby::Client.new(
      client_access_token: '[YOU CLIENT ACCESS TOKEN]',
      api_session_id: 0
    )
  end

  def run(command)
    return @client.text_request(command)[:result][:fulfillment][:speech]
  end

end
