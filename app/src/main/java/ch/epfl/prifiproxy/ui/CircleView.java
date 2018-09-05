package ch.epfl.prifiproxy.ui;

import android.content.Context;
import android.graphics.Canvas;
import android.graphics.DashPathEffect;
import android.graphics.Paint;
import android.os.Build;
import android.support.annotation.Nullable;
import android.util.AttributeSet;
import android.view.View;

import ch.epfl.prifiproxy.R;

public class CircleView extends View {
    private final Paint paint;
    private final float STROKE_WIDTH = 4f;
    private final float DASH_DISTANCE = 8f;

    public CircleView(Context context, @Nullable AttributeSet attrs) {
        super(context, attrs);
        this.paint = new Paint();

        int colorId = R.color.colorOn;

        int color;
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            color = context.getColor(colorId);
        } else {
            color = context.getResources().getColor(colorId);
        }

        float f = DASH_DISTANCE * getResources().getDisplayMetrics().density;
        DashPathEffect dashPath = new DashPathEffect(new float[]{f, f}, 0);

        paint.setPathEffect(dashPath);


        paint.setColor(color);
        paint.setStyle(Paint.Style.STROKE);
        paint.setStrokeWidth(STROKE_WIDTH);
        paint.setAntiAlias(true);
    }


    @Override
    protected void onDraw(Canvas canvas) {
        super.onDraw(canvas);
        float radius = (canvas.getWidth() / 2) - STROKE_WIDTH;
        int posX = canvas.getWidth() / 2;
        int posY = canvas.getHeight() / 2;
        canvas.drawCircle(posX, posY, radius, paint);
    }
}