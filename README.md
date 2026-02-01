# cl-assignment

 
##	Task1 Manual set up of docker and 

 
	
	#Docker script execute with shared folder to "extract" artifacts-rpms.
	docker run -it --rm \
	  -u 0 \
	  -v "$PWD:/work" \   #visible in container under /work
	  stream8-kernel-builder \
	  /bin/bash
##	manual-instructions inside container  

	#once on the Docker VM.
	[root@764ba3dcf6c8 work]# history
	      cd /work/
          rpmdev-setuptree
	      rpm -ivh kernel-4.18.0-448.el8.src.rpm 
	      dnf -y builddep /root/rpmbuild/SPECS/kernel.spec
	      rpmbuild -bb /root/rpmbuild/SPECS/kernel.spec
		#verify rpm's are actually there.
	      ls /root/rpmbuild/RPMS/x86_64/*.rpm  
		 #prepere directory on host.
	      mkdir -p /work/out/rpms   
		#move to host.
	      cp -av /root/rpmbuild/RPMS/x86_64/* /work/out/rpms/   

After completion we have the RPMs copied back to the host machine in $PWD/out/rpms.

Change permissions to user's on host:

	sudo chown -R $(id -u):$(id -g) out/rpms

 
## task2 Golang automation for creation of RPMs.
 
    mkdir cmd
    vi cmd/main.go
    mkdir -p cmd/build-stream8-kernel
    vi cmd/build-stream8-kernel/go.mod
    go build -o build-stream8-kernel ./cmd/build-stream8-kernel
    mkdir -p out_go_test
	##example using the vanilla src.rpm and output on out_go_test directory.
    ./build-stream8-kernel ./kernel-4.18.0-448.el8.src.rpm ./out_go_test/
	#example usage using the same srpm but from provided URL
    mkdir -p out_go_url
    ./build-stream8-kernel  "https://vault.centos.org/8-stream/BaseOS/Source/SPackages/kernel-4.18.0-448.el8.src.rpm" ./out_go_url

 
## Task3 Patches application using the Go tool.
 Task3 produces only a patched SRPM using rpmbuild -bs. Binary RPMs are generated later exclusively via the Go automation tool.
 
	#run container and prepare SPRM
    rpmdev-setuptree
	rpm -ivh kernel-4.18.0-448.el8.src.rpm 
	dnf -y builddep /root/rpmbuild/SPECS/kernel.spec
	cd /root/rpmbuild/SOURCES
	curl -L -o 80e648042e512d5a767da251d44132553fe04ae0.patch  "https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/patch/?id=80e648042e512d5a767da251d44132553fe04ae0"
	curl -L -o f90fff1e152dedf52b932240ebbd670d83330eca.patch  "https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/patch/?id=f90fff1e152dedf52b932240ebbd670d83330eca"

	##modify /root/rpmbuild/SPECS/kernel.spec to include the patches.
	## add Patch1000000, and Patch1000001 just before the test-patch
	##  as 
	##		   Patch1000000: 80e648042e512d5a767da251d44132553fe04ae0.patch
	##		   Patch1000001: f90fff1e152dedf52b932240ebbd670d83330eca.patch
	##
	## and also adding before the ApplyOptionalPatch linux-kernel-test.patch the followings:
	##
	##		   ApplyPatch 80e648042e512d5a767da251d44132553fe04ae0.patch
	##		   ApplyPatch f90fff1e152dedf52b932240ebbd670d83330eca.patch    	
	## then:: 

	rpmbuild -bb /root/rpmbuild/SPECS/kernel.spec  # this will fail with error.
	#previous fails, 
    #since rpmbuild -bb failed we use the resulting BUILD tree to edit and regenerate the patch

    cd /root/rpmbuild/BUILD/kernel-4.18.0-448.el8/linux-4.18.0-448.el8.x86_64
	git init  > /dev/null 
	git add -A
	git commit -m baseline > /dev/null

	#manual modifications for patch needing:
		#diff --git a/kernel/time/posix-cpu-timers.c b/kernel/time/posix-cpu-timers.c
		index bd856eee9..1f956316b 100644
		--- a/kernel/time/posix-cpu-timers.c
		+++ b/kernel/time/posix-cpu-timers.c
		@@ -1121,7 +1121,9 @@ void run_posix_cpu_timers(void)
			LIST_HEAD(firing);
		 
			lockdep_assert_irqs_disabled();
		-
		+       //patched manual in the file.(TASK3)
		+       if(tsk->exit_state)
		+               return;
			/*
			 * The fast path checks that there are no expired thread or thread
			 * group timers.  If that's so, just return.
	

	git diff > ~/rpmbuild/SOURCES/f90fff1e152dedf52b932240ebbd670d83330eca.patch ##overwrite it to work
    
    cd /root/rpmbuild
	#getting the SRPM
	rpmbuild --bs ~/rpmbuild/SPECS/kernel.spec --define "buildid .patchedTASK3"
	cp ~/rpmbuild/SRPMS/kernel-4.18.0-448.el8.patchedTASK3.src.rpm /work

	#then use the Go tool to produce in patched_rpms folder the binaries.
	./build-stream8-kernel kernel-4.18.0-448.el8.patchedTASK3.src.rpm patched_rpms
	

# optional-add on (shell script) for verification.
 After tool finishes we can run verify to inspect
 source code via GDB 
 .e,g. list run_posix_cpu_timers or list mac_partition
 and inspect inside the source code that changes are applied.	
	
	./verify.sh

  
  We can see our patched function code in the following figure.
  ![img](https://github.com/NikosMouzakitis/cl-assignment/blob/main/png/verification.png)

