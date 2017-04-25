#!/usr/bin/ruby -w

require 'colorize'
require 'api-ai-ruby'

require './lib/ai'
require './lib/interpreter'

class Amsh

  def initialize
    @ai = Ai.new
    @interpreter = Interpreter.new

    puts "\e[H\e[2J"
    puts "amsh v0.1b"
    puts "APE MACHINE SHELL - Coded by: Daniel Owen van Dommelen"
    puts "\nREADY"
  end

  def run
    loop do
      printf '>'
      @cmd = gets.gsub("\n", '')

      @response    = @ai.run(@cmd)
      @interpreted = @interpreter.run(@response, @cmd)

      puts @interpreted.yellow
    end
  end

end

amsh = Amsh.new
amsh.run
