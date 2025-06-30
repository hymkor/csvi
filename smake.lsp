(defglobal make (load "smake-go120.lsp"))

(defun build ()
  (funcall make 'build)
  (funcall make 'build-cmd "cmd/csvi"))

(case $1
  (("dist")
   (dolist (platform
             '(("darwin"  "amd64")
               ("darwin"  "arm64")
               ("linux"   "386")
               ("linux"   "amd64")
               ("windows" "386" )
               ("windows" "amd64")))
     (env
       (("GOOS"   (car platform))
        ("GOARCH" (car (cdr platform))))
       (let ((exe-files (build)))
         (funcall make 'dist exe-files))
       ) ; env
     ) ; dolist
   ) ; "dist"

  (("build" "" nil)
   (build)
   ) ; "build

  (t
    (funcall make $1)
    ) ; t

  ) ; case
