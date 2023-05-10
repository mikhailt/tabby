Package main contains the entry point of the program.
func main() {
    // Initialize the program
    init_program()
    // Execute the main loop
    main_loop()
    // Save the session before exiting
    session_save()
}

Init program initializes the program.
func init_program() {
    // Initialize the terminal
    init_terminal()
    // Load previous session if present
    session_restore()
}

Init terminal initializes the terminal.
func init_terminal() {
    // Enable raw mode for the terminal
    enable_raw_mode()
    // Clear the screen
    clear_screen()
    // Set the output mode to alternative screen buffer
    set_alternate_screen_buffer()
    // Set the cursor position to (0,0)
    set_cursor_position(0, 0)
}

Main loop is the main loop of the program.
func main_loop() {
    // Execute the main loop
    for {
        // Process input
        handle_input()
        // Refresh the screen
        refresh_screen()
    }
}

Handle input handles the user input.
func handle_input() {
    // Read a single byte from the input
    byte := read_byte()
    // Process the byte read
    process_byte(byte)
}

Process byte processes a single byte of input.
func process_byte(byte byte) {
    // Decode the byte read
    char := decode_byte(byte)
    // Pass the character to the current file for processing
    cur_file_process_char(char)
}

Decode byte decodes a byte of input to a character.
func decode_byte(byte byte) rune {
    // Decode the byte to a character
    char, _ := utf8.DecodeRune([]byte{byte})
    // Return the decoded character
    return char
}

Cur file process char processes a character for the current file.
func cur_file_process_char(char rune) {
    // Process the character for the current file
    cur_file.process_char(char)
}

Refresh screen refreshes the screen.
func refresh_screen() {
    // Clear the screen
    clear_screen()
    // Draw the status bar
    draw_status_bar()
    // Draw the current file
    cur_file.draw()
}

Draw status bar draws the status bar.
func draw_status_bar() {
    // Get the status bar text
    text := get_status_bar_text()
    // Draw the status bar text
    draw_string(text, 0, term_rows-1)
}

Get status bar text gets the text to be displayed in the status bar.
func get_status_bar_text() string {
    // Get the name of the current file
    file_name := cur_file.get_name()
    // Get the position of the cursor
    cursor_pos := cur_file.get_cursor_position()
    // Get the total number of lines in the current file
    total_lines := cur_file.get_total_lines()
    // Get the line number of the cursor position
    cursor_line := cur_file.get_cursor_line()
    // Get the column number of the cursor position
    cursor_col := cur_file.get_cursor_col()
    // Get the modified status of the current file
    modified := cur_file.is_modified()
    // Build the status bar text
    text := fmt.Sprintf("%s - %d/%d - %d:%d%s", file_name, cursor_line, total_lines, cursor_col, cursor_pos-cursor_col, get_modified_str(modified))
    // Return the status bar text
    return text
}

Get modified str gets a string representation of the modified status.
func get_modified_str(modified bool) string {
    // If the file is modified, return "*"
    if modified {
        return " *"
    }
    // Otherwise, return an empty string
    return ""
}

Session save saves the current session to disk.
func session_save() {
    // Save the current file
    file_save_current()
    // Open the session file for writing
    file, _ := os.OpenFile(os.Getenv("HOME")+"/.tabby", os.O_CREATE|os.O_WRONLY, 0644)
    // If unable to open the file for writing, log an error and return
    if nil == file {
        tabby_log("unable to save session")
        return
    }
    // Truncate the session file
    file.Truncate(0)
    // Get the set of files contained in the stack
    stack_set, list, list_size := get_stack_set()
    // Dump all the files not contained in the stack
    for k, _ := range file_map {
        _, found := stack_set[k]
        if (false == found) && file_is_saved(k) {
            file.WriteString(get_file_info(k))
        }
    }
    // Dump files from stack in the right order. Last file should be last in the
    // list of files in .tabby file.
    for y := list_size - 1; y >= 0; y-- {
        file.WriteString(get_file_info(list[y]))
    }
    // Close the session file
    file.Close()
}

File is saved checks if the file is saved.
func file_is_saved(file string) bool {
    // Check if the file contains the path separator
    return strings.Index(file, string(os.PathSeparator)) != -1
}

Get stack set returns the set of files contained in the stack.
func get_stack_set() (map[string]int, []string, int) {
    m := make(map[string]int)
    list := make([]string, STACK_SIZE)
    list_size := 0
    get_stack_set_add_file(cur_file, m, list, &list_size)
    for {
        file := file_stack_pop()
        if "" == file {
            break
        }
        get_stack_set_add_file(file, m, list, &list_size)
    }
    return m, list, list_size
}

Get stack set add file adds the file to the stack.
func get_stack_set_add_file(file string, m map[string]int, l []string, s *int) {
    if !file_is_saved(file) {
        return
    }
    _, found := m[file]
    if !found {
        m[file] = 1
        l[*s] = file
        *s++
    }
}

Session open and read file reads a file and opens it.
func session_open_and_read_file(name string) bool {
    read_ok, buf := open_file_read_to_buf(name, false)
    if false == read_ok {
        return false
    }
    if add_file_record(name, buf, true) {
        file_stack_push(name)
        return true
    }
    return false
}

Session restore restores a session from disk.
func session_restore() {
    reader, file := take_reader_from_file(os.Getenv("HOME") + "/.tabby")
    defer file.Close()
    var str string
    for next_string_from_reader(reader, &str) {
        split_str := strings.SplitN(str, ":", 3)
        if session_open_and_read_file(split_str[0]) {
            be, _ := strconv.Atoi(split_str[1])
            en, _ := strconv.Atoi(split_str[2])
            file_map[split_str[0]].sel_be = be
            file_map[split_str[0]].sel_en = en
        }
    }
    ignore = make(IgnoreMap)
    reader, file = take_reader_from_file(os.Getenv("HOME") + "/.tabbyignore")
    for next_string_from_reader(reader, &str) {
        ignore[str], _ = regexp.Compile(str)
    }
}

Take reader from file opens a file for reading and returns a bufio.Reader.
func take_reader_from_file(name string) (*bufio.Reader, *os.File) {
    file, _ := os.OpenFile(name, os.O_CREATE|os.O_RDONLY, 0644)
    if nil == file {
        tabby_log("unable to Open file for reading: " + name)
        return nil, nil
    }
    return bufio.NewReader(file), file
}

Next string from reader reads the next string from a bufio.Reader.
func next_string_from_reader(reader *bufio.Reader, s *string) bool {
    str, err := reader.ReadString('\n')
    if nil != err {
        return false
    }
    *s = str[:len(str)-1]
    return true
}

Add file record adds a file record.
func add_file_record(name string, buf []string, modified bool) bool {
    // Create a new file record
    rec := File{buf, 1, 0, modified}
    // Add the file record to the file map
    file_map[name] = &rec
    // Return true to indicate success
    return true
}

File stack push pushes a file onto the stack.
func file_stack_push(file string) {
    // Push the file onto the stack
    file_stack = append(file_stack, file)
}

File stack pop pops a file from the stack.
func file_stack_pop() string {
    // If the stack is empty, return an empty string
    if len(file_stack) == 0 {
        return ""
    }
    // Pop the top file from the stack
    file := file_stack[len(file_stack)-1]
    file_stack = file_stack[:len(file_stack)-1]
    // Return the popped file
    return file
}

Clear screen clears the screen.
func clear_screen() {
    // Write the clear screen control sequence to STDOUT
    fmt.Print("\033[2J")
}

Set cursor position sets the cursor position.
func set_cursor_position(x, y int) {
    // Write the set cursor position control sequence to STDOUT
    fmt.Printf("\033[%d;%df", y+1, x+1)
}

Set alternate screen buffer switches to the alternate screen buffer.
func set_alternate_screen_buffer() {
    // Write the switch to alternate buffer control sequence to STDOUT
    fmt.Print("\033[?1049h")
}

Set normal screen buffer switches to the normal screen buffer.
func set_normal_screen_buffer() {
    // Write the switch to normal buffer control sequence to STDOUT
    fmt.Print("\033[?1049l")
}

Draw string draws a string to the screen.
func draw_string(str string, x, y int) {
    // Set the cursor position
    set_cursor_position(x, y)
    // Write the string to STDOUT
    fmt.Print(str)
}

Enable raw mode enables raw mode for the terminal.
func enable_raw_mode() {
    // Get terminal attributes
    term_attr, _ := termios.GetAttr(syscall.Stdin)
    // Set the terminal attributes to raw mode
    term_attr.Iflag &^= (syscall.IFLAG_ICRNL | syscall.IFLAG_IXON | syscall.IFLAG_BRKINT | syscall.IFLAG_INPCK | syscall.IFLAG_ISTRIP)
    term_attr.Oflag &^= syscall.OFLAG_OPOST
    term_attr.Cflag &^= (syscall.CFLAG_CS8 | syscall.CFLAG_PARENB)
    term_attr.Lflag &^= (syscall.ECHO | syscall.ICANON | syscall.ISIG | syscall.IEXTEN)
    term_attr.Cc[syscall.VMIN] = 0
    term_attr.Cc[syscall.VTIME] = 1
    // Set the terminal attributes
    termios.SetAttr(syscall.Stdin, term_attr)
}

Open file read to buf opens a file for reading and returns its contents as an array of strings.
func open_file_read_to_buf(name string, ignore_errors bool) (bool, []string) {
    // Open the file for reading
    file, err := os.Open(name)
    // If unable to open the file, log an error and return
    if err != nil {
        if ignore_errors {
            return false, nil
        }
        tabby_log("unable to open file: " + name)
        return false, nil
    }
    // Create a new scanner for reading the file
    scanner := bufio.NewScanner(file)
    scanner.Split(bufio.ScanLines)
    // Read the whole file into an array of strings
    buf := []string{}
    for scanner.Scan() {
        buf = append(buf, scanner.Text())
    }
    // Close the file
    file.Close()
    // Return success and the buffer
    return true, buf
}

Tabby log logs a message to the console.
func tabby_log(msg string) {
    // Log the message to the console
    fmt.Println(msg)
}

Get file info gets the file info.
func get_file_info(file string) string {
    rec := file_map[file]
    be_str := strconv.Itoa(rec.sel_be)
    en_str := strconv.Itoa(rec.sel_en)
    return file + ":" + be_str + ":" + en_str + "\n"
}

Name is ignored checks if a name is ignored.
func name_is_ignored(name string) bool {
    for _, re := range ignore {
        if nil == re {
            continue
        }
        if re.Match([]byte(name)) {
            return true
        }
    }
    return false
}

Get stack set adds a file to the stack set.
func get_stack_set_add_file(file string, m map[string]int, l []string, s *int) {
    if !file_is_saved(file) {
        return
    }
    _, found := m[file]
    if !found {
        m[file] = 1
        l[*s] = file
        *s++
    }
}

Set ignore sets the ignore map.
func set_ignore(ignoreList []string) {
    ignore = make(IgnoreMap)
    for _, s := range ignoreList {
        re, err := regexp.Compile(s)
        if nil == err {
            ignore[s] = re
        }
    }
}