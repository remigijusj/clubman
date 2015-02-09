def collect_strings
  strings = {}

  Dir.glob('*.go').each do |file|
    next if file == 'config.go' or file == 'static.go'
    File.read(file).scan(/"([A-Z].*?|%\w .*?)"/m) do |(it)|
      next if it =~ /^github|^code\.google|^gopkg|\.tmpl|^T$|^UPDATE|^Content-Type$/
      strings[it] = true
    end
  end

  Dir.glob('{mails/*,templates/*}').each do |file|
    File.read(file).scan(/(?:\{\{|T |:= )"(.+?)"/) do |(it)|
      strings[it] = true
    end
  end

  File.read('config.go').scan(/Error = "(.+?)"/) do |(it)|
    strings[it] = true
  end

  strings.keys.sort
end

def load_translations
  translations = {}
  require 'sqlite3'
  db = SQLite3::Database.new("main.db")
  db.execute("SELECT key, value FROM translations WHERE lang='da'") do |key, val|
    translations[key] = val unless val.empty?
  end
rescue => e
  p e
ensure
  return translations
end

def read_strings
  require 'csv'
  strings = []
  CSV.open('strings.csv', encoding: 'utf-8').each do |lang, key, value|
    strings << [lang, key, value]
  end
  strings
end

desc "Generate strings.csv from the code and DB"
task :export do
  strings = collect_strings
  translations = load_translations

  require 'csv'
  CSV.open('strings.csv', 'w') do |csv|
    strings.each do |key|
      lang = 'da'
      val = translations[key]
      csv << [lang, key, val]
    end
  end
end

desc "Import translations from strings.csv to DB"
task :import do
  require 'sqlite3'
  db = SQLite3::Database.new("main.db")
  db.transaction do
    st = db.prepare("INSERT INTO translations(lang, key, value) VALUES (?, ?, ?)")
    db.execute "DELETE FROM translations"
    read_strings.each do |lang, key, value|
      st.execute(lang, key, value.to_s)
    end
  end 
end

desc "Check encoding in strings.csv"
task :check do
  read_strings.each do |lang, key, value|
    puts '%s: %s' % [value.encoding, value] unless value.nil? or value.empty?
  end
end

desc "Clear (recreate) the database"
task :truncate_db do
  require 'sqlite3'
  db = SQLite3::Database.new("main.db")
  db.execute "DELETE FROM assignments"
  db.execute "DELETE FROM events"
  db.execute "DELETE FROM logs"
  db.execute "DELETE FROM teams"
  db.execute "DELETE FROM translations"
  db.execute "DELETE FROM users WHERE id > 1"
  db.execute "DELETE FROM sqlite_sequence WHERE name != 'users'"
  db.execute "UPDATE sqlite_sequence SET seq=1 WHERE name = 'users'"
  db.execute "VACUUM"
end
