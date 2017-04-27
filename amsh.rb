#!/usr/bin/ruby -w

require 'colorize'
require 'api-ai-ruby'
require 'tts'

require './lib/ai'
require './lib/interpreter'

class Amsh

  def initialize
    @interpreter = Interpreter.new

    puts "\e[H\e[2J"
    puts "amsh v0.1b"
    puts "APE MACHINE SHELL - Coded by: Daniel Owen van Dommelen"
    puts "\nREADY...\n\n"
  end

  def run
    loop do
      printf "AM>"
      @cmd = gets.gsub("\n", '')

      @response = @interpreter.run(@response, @cmd)
      puts @response.yellow

      Thread.new do
        @response.play
      end
    end
  end

end

$VERBOSE = nil

amsh = Amsh.new
amsh.run
