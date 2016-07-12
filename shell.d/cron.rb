#!/usr/bin/env ruby
crons = []

Dir['/var/spool/cron/*'].each do |crontab|
  user = File.basename(crontab)
  entry = {}

  File.read(crontab).lines.each do |line|
    case line
    when /^([\#]\s*)?([\d\w\,\*\/]+)\s+([\d\,]+|\*)\s+([\d\,]+|\*)\s+([\d\,]+|\*)\s+([\d\,]+|\*)\s+(.*)$/
      entry = entry.merge({
        :user     => user,
        :filename => crontab,
        :disabled => $~[1].to_s.include?('#'),
        :minute   => $~[2],
        :hour     => $~[3],
        :day      => $~[4],
        :month    => $~[5],
        :weekday  => $~[6],
        :command  => $~[7],
      })
    when /^([\w\d\_]+)="?(.*)"?$/
      entry[:env] ||= {}
      entry[:env][$~[1]] = $~[2]
    end
  end

  crons << entry unless entry.empty?
end

Dir['/etc/cron.d/*'].each do |crontab|
  crond_name = File.basename(crontab)
  entry = {}

  File.read(crontab).lines.each do |line|
    case line
    when /^([\#]\s*)?([\d\w\,\*\/]+)\s+([\d\,]+|\*)\s+([\d\,]+|\*)\s+([\d\,]+|\*)\s+([\d\,]+|\*)\s+(\w+)\s+(.*)$/
      entry = entry.merge({
        :crond    => crond_name,
        :filename => crontab,
        :disabled => $~[1].to_s.include?('#'),
        :user     => $~[7],
        :minute   => $~[2],
        :hour     => $~[3],
        :day      => $~[4],
        :month    => $~[5],
        :weekday  => $~[6],
        :command  => $~[8],
      })
    when /^([\w\d\_]+)="?(.*)"?$/
      entry[:env] ||= {}
      entry[:env][$~[1]] = $~[2]
    end
  end

  crons << entry unless entry.empty?
end


crons.each_index do |i|
  cron = crons[i]

  cron.each do |field, value|
    next if value == '*'

    case field.to_sym
    when :disabled
      puts "cron.#{ i }.#{ field }:bool:#{ value }"
    when :env
      value.each do |name, v|
        puts "cron.#{ i }.env.#{ name }:str:#{ v }"
      end
    else
      puts "cron.#{ i }.#{ field }:str:#{ value }"
    end

  end
end
