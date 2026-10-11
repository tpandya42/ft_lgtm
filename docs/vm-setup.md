# VM Setup Guide

This guide explains how to install **Ubuntu 24.04 LTS Desktop** in Oracle VirtualBox, install the VirtualBox Guest Additions, and configure shared folders to access files from your host computer.

The main goal is not just to create a virtual machine. This guide exists because the VM itself is straightforward, but getting the VirtualBox shared folder working on the school computers took much longer to figure out than expected. The important part is making that shared folder reliable, so Ubuntu can access and edit the project files from the host system.

```text
Host project folder, for example ft_lgtm
        |
        | 1. Create an Ubuntu VM
        |    Why: gives you the Linux environment for the project
        v
Ubuntu 24.04 running in VirtualBox
        |
        | 2. Install required packages
        |    Why: lets Guest Additions build correctly
        v
VirtualBox Guest Additions
        |
        | 3. Add your user to the vboxsf group and reboot
        |    Why: gives Ubuntu permission to open shared folders
        v
Shared-folder permissions are active
        |
        | 4. Configure the shared folder in VirtualBox
        |    Why: connects the host project folder to Ubuntu
        v
/media/ft_lgtm
        |
        | 5. Verify that files can be opened, edited, and created
        |    Why: confirms the shared folder actually works
        v
Working project folder inside Ubuntu
```

## 1. Install Ubuntu 24.04 Desktop

### Download Ubuntu

1. Open https://ubuntu.com/download/desktop
2. Download **Ubuntu 24.04 LTS Desktop (64-bit)**.
3. Save the `.iso` file on your host computer.

**Important:** Use the Desktop version, not Ubuntu Server. The Desktop version includes a graphical user interface.

### Create a Virtual Machine

1. Open Oracle VirtualBox.
2. Click **New**.
3. Enter a name for your VM, for example `lgtm`.
4. Select the downloaded Ubuntu ISO file.
5. Select Linux / Ubuntu (64-bit), if required.
6. Configure the virtual machine with approximately:
   - RAM: 9 GB 
   - CPU: 9 cores 
   - Virtual hard disk: 30 GB minimum
7. Complete the VirtualBox setup wizard.
8. Start the virtual machine.
9. Follow the Ubuntu installer and choose **Install Ubuntu**.
10. Create your Ubuntu username and password.
11. Finish the installation and restart Ubuntu.

You should now see the Ubuntu desktop.

## 2. Install Required Packages

Before installing VirtualBox Guest Additions, Ubuntu needs a few additional packages.

**This step is important!** Without these packages, the Guest Additions installation may fail.

Open the Ubuntu terminal by pressing:

`Ctrl + Alt + T`

Copy and paste this command:

```bash
sudo apt update
sudo apt install -y bzip2 tar build-essential dkms linux-headers-$(uname -r)
```

Press Enter.

Ubuntu may ask for your password. Type the password you created during installation and press Enter.

**Note:** Your password will not be displayed while typing. This is normal.

Wait until the command finishes.

The packages include:

- `bzip2` — required to extract VirtualBox Guest Additions
- `tar` — archive extraction utility
- `build-essential` — compiler and build tools
- `dkms` — helps maintain kernel modules
- `linux-headers` — required for building modules for your current Linux kernel

## 3. Install VirtualBox Guest Additions

Guest Additions allow Ubuntu to integrate properly with VirtualBox, including automatic shared-folder mounting.

### Insert the Guest Additions CD

1. Make sure your Ubuntu virtual machine is running.
2. In the VirtualBox menu at the top, click:

   **Devices → Insert Guest Additions CD Image**

3. Ubuntu should display the virtual CD in the Files application.
4. Open the mounted CD, usually named something like `VBox_GAs_7.1.18`.
5. If Ubuntu offers **Run Software**, click it and confirm.
6. Enter your password if requested.
7. Wait until the installation finishes.

The version number may be different depending on your VirtualBox installation.

### If the installer reports an error

For example:

```text
bzip2 not found. Please install: bzip2 tar; and try again.
```

Return to Step 2, install the required packages, and run the Guest Additions installer again.

If the graphical installer does not start, open a terminal and run:

```bash
cd /media/$USER/VBox_GAs_*
sudo sh ./VBoxLinuxAdditions.run
```

### Restart Ubuntu

After successful installation:

```bash
sudo reboot
```

## 4. Enable Shared Folder Permissions (Important!)

**Do not skip this step!**

Even if VirtualBox Guest Additions are installed correctly and your shared folder is configured, Ubuntu may prevent you from accessing its contents.

This happens because your Ubuntu user needs permission to access VirtualBox shared folders.

### Step 1: Add Your User to the VirtualBox Group

Open the Ubuntu terminal using `Ctrl + Alt + T`.

Run the following command:

```bash
sudo usermod -aG vboxsf "$USER"
```

Enter your Ubuntu password if requested.

This command adds your current Ubuntu user to the `vboxsf` group, which grants access to VirtualBox shared folders.

### Step 2: Restart Ubuntu

**This is required!** Group membership changes do not take effect in your current login session.

```bash
sudo reboot
```

Wait for Ubuntu to restart and log in again.

### Step 3: Verify Your Permissions

Open the terminal and run:

```bash
groups
```

You should see `vboxsf` somewhere in the output.

For example:

```text
youruser adm cdrom sudo dip plugdev users lpadmin vboxsf
```

If `vboxsf` appears, your user has the correct group membership.

You can now continue with **Step 5: Create a Shared Folder in VirtualBox**.

**Note:** Without this step, Ubuntu may display a "Permission denied" error when you try to open a shared folder, even if the folder is mounted successfully.

## 5. Create a Shared Folder in VirtualBox

Now we will share a folder from your host computer with Ubuntu.

For example, assume your project folder on the host computer is:

```text
/home/your-host-user/ft_lgtm
```

Replace `your-host-user` with your actual host username.

### Open Shared Folder Settings

1. Start your Ubuntu virtual machine.
2. In the VirtualBox menu, click:

   **Devices → Shared Folders → Shared Folders Settings**

3. Click the **folder icon with a green plus sign** to add a new shared folder.

### Configure the Shared Folder

Use these settings:

| Setting | Value |
|---|---|
| Folder Path | Select your project folder on the host |
| Folder Name | `ft_lgtm` |
| Mount Point | Leave empty |
| Read-only | Unchecked |
| Auto-mount | Checked |
| Make Permanent | Checked |

**Explanation:**

- **Folder Path:** The actual directory on your host computer, not inside Ubuntu.
- **Folder Name:** The name VirtualBox uses to identify this shared folder.
- **Mount Point:** Leave this empty so VirtualBox can choose its standard location.
- **Read-only:** Keep unchecked if you want to edit or create files from Ubuntu.
- **Auto-mount:** Automatically mounts the folder when Ubuntu starts.
- **Make Permanent:** Keeps the shared folder configuration after restarting the virtual machine.

Click **OK** to save the new shared folder.

Click **OK** again to close the VirtualBox settings.

## 6. Access the Shared Folder in Ubuntu

Restart Ubuntu if you have not done so since installing Guest Additions or changing group membership.

VirtualBox normally mounts shared folders under:

```text
/media/<folder_name>
```

In our example:

```text
/media/ft_lgtm
```

### Open the Folder Graphically

1. Open the **Files** application in Ubuntu (the folder icon in the left dock).
2. Press `Ctrl + L` to enter a location.
3. Type:

   ```text
   /media/ft_lgtm
   ```

4. Press Enter.

You should now see the files from your host computer.

You can open, edit, create, and save files directly inside this folder.

**Congratulations! Your shared folder is working.**

### Optional: Add a Bookmark

For easier access in the future:

1. Open `/media/ft_lgtm` in the Files application.
2. Press `Ctrl + D` to bookmark the location.
3. The folder should now appear in the Files sidebar.

You can access it without typing the path again.

## 7. Verify the Shared Folder Using the Terminal

This step is optional.

Open a terminal and run:

```bash
ls /media/
```

You should see something like:

```text
ft_lgtm
```

List the files in your shared folder:

```bash
ls -la /media/ft_lgtm
```

To work directly inside the folder:

```bash
cd /media/ft_lgtm
```

The same files are available from both the host and Ubuntu.

## 8. Troubleshooting

### Problem: The shared folder does not appear

First, check:

- Guest Additions were successfully installed.
- Auto-mount is enabled in VirtualBox.
- Make Permanent is enabled.
- The host folder exists.
- Ubuntu was restarted after installing Guest Additions.

Check whether the shared-folder module is available:

```bash
sudo modprobe vboxsf
```

No output usually indicates that the command succeeded.

Restart Ubuntu and try opening the folder again.

### Problem: Permission denied

Check your groups:

```bash
groups
```

Look for `vboxsf`.

If it is missing:

```bash
sudo usermod -aG vboxsf "$USER"
sudo reboot
```

Try opening the folder after restarting.

### Problem: Guest Additions installation fails

Make sure all dependencies are installed:

```bash
sudo apt update
sudo apt install -y bzip2 tar build-essential dkms linux-headers-$(uname -r)
```

Insert the Guest Additions CD again and rerun the installer.

### Problem: The folder is not mounted automatically

If the folder does not appear even after checking the settings and Guest Additions, you can test manual mounting:

```bash
sudo mkdir -p /mnt/ft_lgtm
sudo mount -t vboxsf ft_lgtm /mnt/ft_lgtm
```

Then check:

```bash
ls /mnt/ft_lgtm
```

If this works, VirtualBox can access the shared folder, but its automatic mounting is not working correctly.

This manual mount is a troubleshooting workaround, not a replacement for fixing automatic mounting.

## 9. Final Checklist

Before you start working on your project, make sure:

- [ ] Ubuntu 24.04 LTS Desktop is installed.
- [ ] All required packages, including `bzip2`, are installed.
- [ ] VirtualBox Guest Additions are installed successfully.
- [ ] Your Ubuntu user belongs to the `vboxsf` group.
- [ ] The shared folder points to the correct host directory.
- [ ] Read-only is disabled.
- [ ] Auto-mount is enabled.
- [ ] Make Permanent is enabled.
- [ ] Ubuntu has been restarted.
- [ ] The folder opens in Ubuntu Files.
- [ ] You can create and edit files inside the shared folder.

---

**Done!**

Your Ubuntu 24.04 virtual machine can now access your host project's files through a VirtualBox shared folder. With automatic mounting enabled and working, the folder should remain available after future restarts without requiring manual mount commands.
