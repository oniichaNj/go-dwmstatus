/* Source: dwm.suckless.org/dwmstatus/getvol.c*/
#include <alsa/asoundlib.h>
#include <alsa/control.h>
int get_volume(void) {
  int vol;
  snd_hctl_t *hctl;
  snd_ctl_elem_id_t *id;
  snd_ctl_elem_value_t *control;
  // To find card and subdevice: /proc/asound/, aplay -L, amixer controls
  snd_hctl_open(&hctl, "hw:0", 0);
  snd_hctl_load(hctl);
  snd_ctl_elem_id_alloca(&id);
  snd_ctl_elem_id_set_interface(id, SND_CTL_ELEM_IFACE_MIXER);
  // amixer controls
  snd_ctl_elem_id_set_name(id, "Master Playback Volume");
  snd_hctl_elem_t *elem = snd_hctl_find_elem(hctl, id);
  snd_ctl_elem_value_alloca(&control);
  snd_ctl_elem_value_set_id(control, id);
  snd_hctl_elem_read(elem, control);
  vol = (int)snd_ctl_elem_value_get_integer(control,0);
  snd_hctl_close(hctl);
  return vol;
}
int get_volume_perc() {
  return 100*get_volume()/74; // 74 is the max number of decibels that will be returned bu get_volume()
}
