import java.io.IOException;

public class TestJava {
    public static void main (String[] arg) throws java.io.IOException {
        int i;
        do {
            System.out.write(i = System.in.read());
        } while (i != -1);
    }
}
