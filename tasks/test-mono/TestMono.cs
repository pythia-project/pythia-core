using System;

namespace Main
{
    public class TestMono {
        public static void Main (string[] args)
        {
            while (true) {
                int input = Console.In.Read();
                if (input == -1) break;
                Console.Out.Write((char)input);
            }
        }
    }
}
