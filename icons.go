package main

var (
iconUnknown = []string{
  "    .-.      ",
  "     __)     ",
  "    (        ",
  "     `-’     ",
  "      •      "}
iconSunny = []string{
  "\033[38;5;226m    \\   /    \033[0m",
  "\033[38;5;226m     .-.     \033[0m",
  "\033[38;5;226m  ― (   ) ―  \033[0m",
  "\033[38;5;226m     `-’     \033[0m",
  "\033[38;5;226m    /   \\    \033[0m"}
iconPartlyCloudy = []string{
  "\033[38;5;226m   \\  /\033[0m      ",
  "\033[38;5;226m _ /\"\"\033[38;5;250m.-.    \033[0m",
  "\033[38;5;226m   \\_\033[38;5;250m(   ).  \033[0m",
  "\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
  "             "}
iconCloudy = []string{
  "             ",
  "\033[38;5;250m     .--.    \033[0m",
  "\033[38;5;250m  .-(    ).  \033[0m",
  "\033[38;5;250m (___.__)__) \033[0m",
  "             "}
iconVeryCloudy = []string{
  "             ",
  "\033[38;5;240;1m     .--.    \033[0m",
  "\033[38;5;240;1m  .-(    ).  \033[0m",
  "\033[38;5;240;1m (___.__)__) \033[0m",
  "             "}
iconLightShowers = []string{
  "\033[38;5;226m _`/\"\"\033[38;5;250m.-.    \033[0m",
  "\033[38;5;226m  ,\\_\033[38;5;250m(   ).  \033[0m",
  "\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
  "\033[38;5;111m     ‘ ‘ ‘ ‘ \033[0m",
  "\033[38;5;111m    ‘ ‘ ‘ ‘  \033[0m"}
iconHeavyShowers = []string{
  "\033[38;5;226m _`/\"\"\033[38;5;240;1m.-.    \033[0m",
  "\033[38;5;226m  ,\\_\033[38;5;240;1m(   ).  \033[0m",
  "\033[38;5;226m   /\033[38;5;240;1m(___(__) \033[0m",
  "\033[38;5;21;1m   ‚‘‚‘‚‘‚‘  \033[0m",
  "\033[38;5;21;1m   ‚’‚’‚’‚’  \033[0m"}
iconLightSnowShowers = []string{
  "\033[38;5;226m _`/\"\"\033[38;5;250m.-.    \033[0m",
  "\033[38;5;226m  ,\\_\033[38;5;250m(   ).  \033[0m",
  "\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
  "\033[38;5;255m     *  *  * \033[0m",
  "\033[38;5;255m    *  *  *  \033[0m"}
iconHeavySnowShowers = []string{
  "\033[38;5;226m _`/\"\"\033[38;5;240;1m.-.    \033[0m",
  "\033[38;5;226m  ,\\_\033[38;5;240;1m(   ).  \033[0m",
  "\033[38;5;226m   /\033[38;5;240;1m(___(__) \033[0m",
  "\033[38;5;255;1m    * * * *  \033[0m",
  "\033[38;5;255;1m   * * * *   \033[0m"}
iconLightSleetShowers = []string{
  "\033[38;5;226m _`/\"\"\033[38;5;250m.-.    \033[0m",
  "\033[38;5;226m  ,\\_\033[38;5;250m(   ).  \033[0m",
  "\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
  "\033[38;5;111m     ‘ \033[38;5;255m*\033[38;5;111m ‘ \033[38;5;255m* \033[0m",
  "\033[38;5;255m    *\033[38;5;111m ‘ \033[38;5;255m*\033[38;5;111m ‘  \033[0m"}
iconThunderyShowers = []string{
  "\033[38;5;226m _`/\"\"\033[38;5;250m.-.    \033[0m",
  "\033[38;5;226m  ,\\_\033[38;5;250m(   ).  \033[0m",
  "\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
  "\033[38;5;228;5m    ⚡\033[38;5;111;25m‘ ‘\033[38;5;228;5m⚡\033[38;5;111;25m‘ ‘ \033[0m",
  "\033[38;5;111m    ‘ ‘ ‘ ‘  \033[0m"}
iconThunderyHeavyRain = []string{
  "\033[38;5;240;1m     .-.     \033[0m",
  "\033[38;5;240;1m    (   ).   \033[0m",
  "\033[38;5;240;1m   (___(__)  \033[0m",
  "\033[38;5;21;1m  ‚‘\033[38;5;228;5m⚡\033[38;5;21;25m‘‚\033[38;5;228;5m⚡\033[38;5;21;25m‚‘   \033[0m",
  "\033[38;5;21;1m  ‚’‚’\033[38;5;228;5m⚡\033[38;5;21;25m’‚’   \033[0m"}
iconThunderySnowShowers = []string{
  "\033[38;5;226m _`/\"\"\033[38;5;250m.-.    \033[0m",
  "\033[38;5;226m  ,\\_\033[38;5;250m(   ).  \033[0m",
  "\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
  "\033[38;5;255m     *\033[38;5;228;5m⚡\033[38;5;255;25m *\033[38;5;228;5m⚡\033[38;5;255;25m * \033[0m",
  "\033[38;5;255m    *  *  *  \033[0m"}
iconLightRain = []string{
  "\033[38;5;250m     .-.     \033[0m",
  "\033[38;5;250m    (   ).   \033[0m",
  "\033[38;5;250m   (___(__)  \033[0m",
  "\033[38;5;111m    ‘ ‘ ‘ ‘  \033[0m",
  "\033[38;5;111m   ‘ ‘ ‘ ‘   \033[0m"}
iconHeavyRain = []string{
  "\033[38;5;240;1m     .-.     \033[0m",
  "\033[38;5;240;1m    (   ).   \033[0m",
  "\033[38;5;240;1m   (___(__)  \033[0m",
  "\033[38;5;21;1m  ‚‘‚‘‚‘‚‘   \033[0m",
  "\033[38;5;21;1m  ‚’‚’‚’‚’   \033[0m"}
iconLightSnow = []string{
  "\033[38;5;250m     .-.     \033[0m",
  "\033[38;5;250m    (   ).   \033[0m",
  "\033[38;5;250m   (___(__)  \033[0m",
  "\033[38;5;255m    *  *  *  \033[0m",
  "\033[38;5;255m   *  *  *   \033[0m"}
iconHeavySnow = []string{
  "\033[38;5;240;1m     .-.     \033[0m",
  "\033[38;5;240;1m    (   ).   \033[0m",
  "\033[38;5;240;1m   (___(__)  \033[0m",
  "\033[38;5;255;1m   * * * *   \033[0m",
  "\033[38;5;255;1m  * * * *    \033[0m"}
iconLightSleet = []string{
  "\033[38;5;250m     .-.     \033[0m",
  "\033[38;5;250m    (   ).   \033[0m",
  "\033[38;5;250m   (___(__)  \033[0m",
  "\033[38;5;111m    ‘ \033[38;5;255m*\033[38;5;111m ‘ \033[38;5;255m*  \033[0m",
  "\033[38;5;255m   *\033[38;5;111m ‘ \033[38;5;255m*\033[38;5;111m ‘   \033[0m"}
iconFog = []string{
  "             ",
  "\033[38;5;251m _ - _ - _ - \033[0m",
  "\033[38;5;251m  _ - _ - _  \033[0m",
  "\033[38;5;251m _ - _ - _ - \033[0m",
  "             "}
)
